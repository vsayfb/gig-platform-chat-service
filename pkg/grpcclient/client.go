package grpcclient

import (
	"context"

	pb "github.com/vsayfb/gig-platform-protos/contracts"
	"google.golang.org/grpc"
)

type UserClient struct {
	client pb.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{
		client: pb.NewUserServiceClient(conn),
	}
}

func (c *UserClient) GetUser(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
	return c.client.GetUser(
		ctx,
		&pb.GetUserRequest{
			UserId: userID,
		},
	)
}

func (c *UserClient) GetUsers(ctx context.Context, userIDs []string) (*pb.GetUsersResponse, error) {
	return c.client.GetUsers(
		ctx, &pb.GetUsersRequest{
			UserIds: userIDs,
		},
	)
}
