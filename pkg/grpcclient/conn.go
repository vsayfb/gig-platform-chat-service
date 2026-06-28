package grpcclient

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCConnection(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)

	defer cancel()

	for {
		state := conn.GetState()

		if state == connectivity.Ready {
			slog.Info("connected to user service")
			return conn, nil
		}

		if !conn.WaitForStateChange(ctx, state) {
			conn.Close()
			return nil, errors.New("timeout connecting to user service")
		}
	}
}
