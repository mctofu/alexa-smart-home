package sqsrelay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mctofu/alexa-smart-home/alexa"
	"github.com/mctofu/alexa-smart-home/deferred"
)

// SQSMessageReader is the subset of sqsiface.SQSAPI used by QueueProcessor
type SQSMessageReader interface {
	ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageWithContext(aws.Context, *sqs.DeleteMessageInput, ...request.Option) (*sqs.DeleteMessageOutput, error)
}

// QueueProcessor reads and handles sqs messages produced by RelayHandler
type QueueProcessor struct {
	SQS                  SQSMessageReader
	QueueURL             string
	Handler              *deferred.Handler
	QueueWaitTimeSeconds int64
}

// Process reads and handles SQS queue messages until an error occurs
func (q *QueueProcessor) Process(ctx context.Context) error {
	for {
		req := sqs.ReceiveMessageInput{
			QueueUrl:        aws.String(q.QueueURL),
			WaitTimeSeconds: aws.Int64(q.QueueWaitTimeSeconds),
		}
		resp, err := q.SQS.ReceiveMessageWithContext(ctx, &req)
		if err != nil {
			return fmt.Errorf("failed to read from sqs: %v", err)
		}

		for _, msg := range resp.Messages {
			var homeReq alexa.Request
			if err := json.Unmarshal([]byte(*msg.Body), &homeReq); err != nil {
				return fmt.Errorf("failed to read message: %s: %v", *msg.Body, err)
			}

			if err := q.Handler.HandleRequest(ctx, &homeReq); err != nil {
				return fmt.Errorf("failed to handle request: %v", err)
			}

			deleteReq := sqs.DeleteMessageInput{
				QueueUrl:      aws.String(q.QueueURL),
				ReceiptHandle: msg.ReceiptHandle,
			}
			if _, err := q.SQS.DeleteMessageWithContext(ctx, &deleteReq); err != nil {
				return fmt.Errorf("failed to delete message: %v", err)
			}
		}
	}
}
