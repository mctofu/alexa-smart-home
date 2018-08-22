package sqsrelay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mctofu/alexa-smart-home/alexa"
)

// SQSMessageSender is the subset of sqsiface.SQSAPI used by RelayHandler
type SQSMessageSender interface {
	SendMessageWithContext(aws.Context, *sqs.SendMessageInput, ...request.Option) (*sqs.SendMessageOutput, error)
}

// RelayHandler publishes the request to a SQS queue
type RelayHandler struct {
	SQS      SQSMessageSender
	QueueURL string
}

// Relay handles the alexa request by marshalling to json and sending it as a SQS message
func (r *RelayHandler) Relay(ctx context.Context, req *alexa.Request) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("sqsrelay: failed to marshal request: %v", err)
	}

	msg := sqs.SendMessageInput{
		MessageBody:            aws.String(string(payload)),
		QueueUrl:               aws.String(r.QueueURL),
		MessageGroupId:         aws.String("alexa.HandleRequest"),
		MessageDeduplicationId: &req.Directive.Header.MessageID,
	}

	_, err = r.SQS.SendMessageWithContext(ctx, &msg)
	if err != nil {
		return fmt.Errorf("sqsrelay: failed to send request to sqs: %v", err)
	}

	return nil
}
