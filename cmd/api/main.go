package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/vsayfb/gig-platform-chat-service/config"
	"github.com/vsayfb/gig-platform-chat-service/hub"
	"github.com/vsayfb/gig-platform-chat-service/internal/message"
	"github.com/vsayfb/gig-platform-chat-service/internal/thread"
	"github.com/vsayfb/gig-platform-chat-service/pkg/database"
	"github.com/vsayfb/gig-platform-chat-service/pkg/grpcclient"
	"github.com/vsayfb/gig-platform-chat-service/pkg/jwt"
	"github.com/vsayfb/gig-platform-chat-service/pkg/logger"
	"github.com/vsayfb/gig-platform-chat-service/pkg/metrics"
	"github.com/vsayfb/gig-platform-chat-service/pkg/middleware"
	sqspkg "github.com/vsayfb/gig-platform-chat-service/pkg/sqs"
	"github.com/vsayfb/gig-platform-chat-service/pkg/tracing"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.AppEnv)

	mongoClient, db, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)

	if err != nil {
		slog.Error("failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.AWSRegion))

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

	conn, err := grpcclient.NewGRPCConnection(cfg.UserServiceGRPCAddr)

	if err != nil {
		slog.Error("failed to create user service client", "err", err)
		os.Exit(1)
	}

	grpcUserClient := grpcclient.NewUserClient(conn)

	threadRepo := thread.NewRepository(db)
	msgRepo := message.NewRepository(db)

	h := hub.New()

	jwtSvc := jwt.New(cfg.JWTSecret)

	wsHandler := thread.NewWSHandler(h, jwtSvc, threadRepo, msgRepo, publisher, grpcUserClient)
	restHandler := thread.NewHandler(threadRepo, msgRepo, grpcUserClient)

	tracingCtx := context.Background()

	shutdownTracer, err := tracing.InitTracer(tracingCtx, cfg.ServiceName, cfg.OTelCollectorAddr)

	if err != nil {
		slog.Error("failed to init tracer", "error", err)
		os.Exit(1)
	}

	metrics.Register()

	metricsSrv := metrics.StartServer(
		cfg.MetricsServerPort,
	)

	r := chi.NewRouter()

	r.Use(cors.AllowAll().Handler)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.MetricsMiddleware)
	r.Use(middleware.TracingMiddleware)

	r.Get("/ws", wsHandler.ServeWS)

	authMiddleware := middleware.Auth(jwtSvc)

	r.With(authMiddleware).Get("/threads", restHandler.ListThreads)
	r.With(authMiddleware).Get("/threads/{threadID}", restHandler.GetThread)
	r.With(authMiddleware).Get("/threads/{threadID}/messages", restHandler.ListMessages)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	serverErr := make(chan error, 1)

	go func() {
		slog.Info(
			"Chat Service listening",
			"port",
			cfg.Port,
		)

		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			slog.Error(
				"HTTP server failed",
				"err",
				err,
			)
		}

	case <-quit:
		slog.Info("shutting down chat service")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error(
			"failed to shutdown HTTP server",
			"err",
			err,
		)
	}

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error(
			"failed to shutdown metrics server",
			"err",
			err,
		)
	}

	if err := conn.Close(); err != nil {
		slog.Error(
			"failed to close grpc connection",
			"err",
			err,
		)
	}

	if err := mongoClient.Disconnect(shutdownCtx); err != nil {
		slog.Error(
			"failed to disconnect MongoDB",
			"err",
			err,
		)
	}

	shutdownTracer(shutdownCtx)

	slog.Info("shutdown complete")
}
