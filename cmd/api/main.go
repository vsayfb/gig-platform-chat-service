package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/vsayfb/gig-platform-chat-service/config"
	"github.com/vsayfb/gig-platform-chat-service/hub"
	"github.com/vsayfb/gig-platform-chat-service/internal/message"
	"github.com/vsayfb/gig-platform-chat-service/internal/thread"
	"github.com/vsayfb/gig-platform-chat-service/pkg/database"
	"github.com/vsayfb/gig-platform-chat-service/pkg/jwt"
	"github.com/vsayfb/gig-platform-chat-service/pkg/logger"
	"github.com/vsayfb/gig-platform-chat-service/pkg/middleware"
	sqspkg "github.com/vsayfb/gig-platform-chat-service/pkg/sqs"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.AppEnv)

	db, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)

	if err != nil {
		slog.Error("failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)

	if err != nil {
		slog.Error("failed to load AWS config", "err", err)
		os.Exit(1)
	}

	sqsClient := sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
		if cfg.SQSEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.SQSEndpoint)
		}
	})

	publisher := sqspkg.NewPublisher(sqsClient, cfg.SQSQueueURL)

	threadRepo := thread.NewRepository(db)
	msgRepo := message.NewRepository(db)

	h := hub.New()

	jwtSvc := jwt.New(cfg.JWTSecret)

	wsHandler := thread.NewWSHandler(h, jwtSvc, threadRepo, msgRepo, publisher)
	restHandler := thread.NewHandler(threadRepo, msgRepo)

	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(middleware.Logger)

	r.Get("/ws", wsHandler.ServeWS)

	authMiddleware := middleware.Auth(jwtSvc)

	r.With(authMiddleware).Get("/threads", restHandler.ListThreads)
	r.With(authMiddleware).Get("/threads/{threadID}", restHandler.GetThread)
	r.With(authMiddleware).Get("/threads/{threadID}/messages", restHandler.ListMessages)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("Chat Service listening", "port", cfg.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "err", err)
			os.Exit(1)
		}

	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("shutting down chat service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_ = srv.Shutdown(ctx)

	slog.Info("shutdown complete")
}
