package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Publisher struct {
	client   *sqs.Client
	queueURL string
}

func NewPublisher(client *sqs.Client, queueURL string) *Publisher {
	return &Publisher{client: client, queueURL: queueURL}
}

type NewMessageEvent struct {
	Type       string `json:"type"`
	ThreadID   string `json:"thread_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	SentAt     string `json:"sent_at"`
}

func (p *Publisher) PublishNewMessage(ctx context.Context, event NewMessageEvent) error {
	event.Type = "new_message"

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("sqs: marshal event: %w", err)
	}

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
	})

	if err != nil {
		return fmt.Errorf("sqs: send message: %w", err)
	}

	return nil
}
