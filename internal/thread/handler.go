package thread

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vsayfb/gig-platform-chat-service/internal/message"
	"github.com/vsayfb/gig-platform-chat-service/pkg/httputil"
	"github.com/vsayfb/gig-platform-chat-service/pkg/middleware"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	threadRepo *Repository
	msgRepo    *message.Repository
}

func NewHandler(threadRepo *Repository, msgRepo *message.Repository) *Handler {
	return &Handler{threadRepo: threadRepo, msgRepo: msgRepo}
}

// GET /threads
func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	threads, err := h.threadRepo.FindByParticipant(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch threads")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, threads)
}

// GET /threads/{threadID}
func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	threadID, err := bson.ObjectIDFromHex(chi.URLParam(r, "threadID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid thread id")
		return
	}

	// Verify the requesting user is a participant
	isParticipant, err := h.threadRepo.IsParticipant(r.Context(), threadID, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify participant")
		return
	}
	if !isParticipant {
		httputil.WriteError(w, http.StatusForbidden, "not a participant in this thread")
		return
	}

	thread, err := h.threadRepo.FindByID(r.Context(), threadID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "thread not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, thread)
}

// GET /threads/{threadID}/messages?cursor=<RFC3339>&limit=50
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	threadID, err := bson.ObjectIDFromHex(chi.URLParam(r, "threadID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid thread id")
		return
	}

	// Verify the requesting user is a participant
	isParticipant, err := h.threadRepo.IsParticipant(r.Context(), threadID, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify participant")
		return
	}
	if !isParticipant {
		httputil.WriteError(w, http.StatusForbidden, "not a participant in this thread")
		return
	}

	// Parse cursor (sent_at of the oldest message client has)
	var cursor time.Time
	if c := r.URL.Query().Get("cursor"); c != "" {
		cursor, err = time.Parse(time.RFC3339, c)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid cursor format, use RFC3339")
			return
		}
	}

	// Parse limit
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	msgs, err := h.msgRepo.FindByThread(r.Context(), threadID, cursor, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch messages")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, msgs)
}
