package thread

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/vsayfb/gig-platform-chat-service/hub"
	"github.com/vsayfb/gig-platform-chat-service/internal/message"
	"github.com/vsayfb/gig-platform-chat-service/pkg/grpcclient"
	"github.com/vsayfb/gig-platform-chat-service/pkg/httputil"
	"github.com/vsayfb/gig-platform-chat-service/pkg/jwt"
	sqspkg "github.com/vsayfb/gig-platform-chat-service/pkg/sqs"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten in production
	},
}

// incomingMessage is what the client sends over WebSocket.
type incomingMessage struct {
	Content string `json:"content"`
}

// outgoingMessage is what we deliver to the recipient.
type outgoingMessage struct {
	ThreadID string `json:"thread_id"`
	SenderID string `json:"sender_id"`
	Content  string `json:"content"`
	SentAt   string `json:"sent_at"`
}

type WSHandler struct {
	hub        *hub.Hub
	jwtSvc     *jwt.Service
	threadRepo *Repository
	msgRepo    *message.Repository
	publisher  *sqspkg.Publisher
	userClient *grpcclient.UserClient
}

func NewWSHandler(
	h *hub.Hub,
	jwtSvc *jwt.Service,
	threadRepo *Repository,
	msgRepo *message.Repository,
	publisher *sqspkg.Publisher,
	userClient *grpcclient.UserClient,
) *WSHandler {
	return &WSHandler{
		hub:        h,
		jwtSvc:     jwtSvc,
		threadRepo: threadRepo,
		msgRepo:    msgRepo,
		publisher:  publisher,
		userClient: userClient,
	}
}

// ServeWS handles WebSocket upgrade and the read loop.
// Query params: token=<jwt>, targetID=<uuid>
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate
	token := r.URL.Query().Get("token")
	if token == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "missing token")
		return
	}

	senderID, err := h.jwtSvc.Verify(token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// 2. Validate targetID
	targetID := r.URL.Query().Get("targetID")
	if targetID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing targetID")
		return
	}

	if senderID == targetID {
		httputil.WriteError(w, http.StatusBadRequest, "cannot connect to yourself")
		return
	}

	// 3. Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws: upgrade failed", "err", err)
		return
	}

	// 4. Register in Hub
	client := &hub.Client{
		UserID:   senderID,
		TargetID: targetID,
		Conn:     conn,
	}
	h.hub.Register(client)

	slog.Info("ws: client connected", "userID", senderID, "targetID", targetID)

	defer func() {
		h.hub.Unregister(senderID)
		conn.Close()
		slog.Info("ws: client disconnected", "userID", senderID)
	}()

	// 5. Read loop
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("ws: read error", "userID", senderID, "err", err)
			}
			break
		}

		var incoming incomingMessage
		if err := json.Unmarshal(raw, &incoming); err != nil {
			slog.Warn("ws: invalid message format", "userID", senderID)
			continue
		}

		if incoming.Content == "" {
			continue
		}

		h.handleMessage(senderID, targetID, incoming.Content)
	}
}

func (h *WSHandler) handleMessage(senderID, targetID, content string) {
	// Use background context since the request context cancels on disconnect
	// but we still want the DB write and SQS publish to complete.
	bgCtx := context.Background()

	// 1. Lazy-create or fetch thread
	t, err := h.threadRepo.FindOrCreate(bgCtx, senderID, targetID, content)
	if err != nil {
		slog.Error("ws: findOrCreate thread failed", "err", err)
		return
	}

	// 2. Persist message
	msg := &message.Message{
		ThreadID: t.ID,
		SenderID: senderID,
		Content:  content,
	}
	if err := h.msgRepo.Insert(bgCtx, msg); err != nil {
		slog.Error("ws: insert message failed", "err", err)
		return
	}

	// 3. Update thread's last_message
	_ = h.threadRepo.UpdateLastMessage(bgCtx, t.ID, content)

	// 4. Build outgoing payload
	outgoing := outgoingMessage{
		ThreadID: t.ID.Hex(),
		SenderID: senderID,
		Content:  content,
		SentAt:   msg.SentAt.Format(time.RFC3339),
	}

	payload, err := json.Marshal(outgoing)
	if err != nil {
		slog.Error("ws: marshal outgoing failed", "err", err)
		return
	}

	// 5. Deliver — WebSocket if online, SQS if offline
	if delivered := h.hub.Send(targetID, payload); !delivered {
		err := h.publisher.PublishNewMessage(bgCtx, sqspkg.NewMessageEvent{
			ThreadID:   t.ID.Hex(),
			SenderID:   senderID,
			ReceiverID: targetID,
			Content:    content,
			SentAt:     msg.SentAt.Format(time.RFC3339),
		})
		if err != nil {
			slog.Error("ws: sqs publish failed", "err", err)
		}
	}
}
