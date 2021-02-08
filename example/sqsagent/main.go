package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mctofu/alexa-smart-home/alexa"
	"github.com/mctofu/alexa-smart-home/aws/s3store"
	"github.com/mctofu/alexa-smart-home/aws/sqsrelay"
	"github.com/mctofu/alexa-smart-home/deferred"
)

// Listens on a SQS queue to remotely handle deferred power controller events
// sent from the skill lambda
func main() {
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	s3TokenBucket := os.Getenv("S3_TOKEN_BUCKET")
	authClientID := os.Getenv("AUTH_CLIENT_ID")
	authClientSecret := os.Getenv("AUTH_CLIENT_SECRET")

	session, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to init aws session: %v", err)
	}

	s3Client := s3.New(session)

	tokenStorage := &alexa.DebugTokenStore{
		TokenStore: &s3store.TokenStorage{
			S3:     s3Client,
			Bucket: s3TokenBucket,
		},
	}

	userIDReader := &alexa.ProfileUserIDReader{HTTPDoer: http.DefaultClient}

	respBuilder := alexa.NewResponseBuilder()

	fanSwitch := fanSwitch{respBuilder}
	windowControl := windowControl{respBuilder}

	mux := alexa.NewNamespaceMux()
	mux.Handle(alexa.NamespacePercentageController,
		alexa.PercentageControllerHandler(
			alexa.HandlerFunc(windowControl.SetPercentage),
			alexa.HandlerFunc(windowControl.AdjustPercentage)))
	mux.Handle(alexa.NamespacePowerController,
		alexa.PowerControllerHandler(
			alexa.HandlerFunc(fanSwitch.TurnOn),
			alexa.HandlerFunc(fanSwitch.TurnOff)))

	requestHandler := mux

	eventSender := &deferred.HTTPEventSender{
		TokenStore:   tokenStorage,
		UserIDReader: userIDReader,
		ClientID:     authClientID,
		ClientSecret: authClientSecret,
	}

	deferredHandler := &deferred.Handler{
		EventSender:    eventSender,
		RequestHandler: alexa.DebugHandler(requestHandler),
	}

	sqsClient := sqs.New(session)

	reader := &sqsrelay.QueueProcessor{
		SQS:                  sqsClient,
		QueueURL:             sqsQueueURL,
		Handler:              deferredHandler,
		QueueWaitTimeSeconds: 20,
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := reader.Process(ctx); err != nil {
				if ctx.Err() != nil {
					log.Printf("Terminating: %v", err)
					break
				}
				log.Printf("Failed to process queue: %v", err)
				delay := time.After(time.Duration(reader.QueueWaitTimeSeconds) * time.Second)
				select {
				case <-delay:
					continue
				case <-ctx.Done():
					continue
				}
			}
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		cancel()
	}

	wg.Wait()
}

type fanSwitch struct {
	respBuilder *alexa.ResponseBuilder
}

func (f fanSwitch) TurnOn(ctx context.Context, req *alexa.Request) (*alexa.Response, error) {
	log.Println("Turn on!")
	return f.respBuilder.BasicResponse(req, alexa.ContextProperty{
		Namespace:                 alexa.NamespacePowerController,
		Name:                      "powerState",
		Value:                     json.RawMessage(`"` + "ON" + `"`),
		TimeOfSample:              time.Now(),
		UncertaintyInMilliseconds: 500,
	}), nil
}

func (f fanSwitch) TurnOff(ctx context.Context, req *alexa.Request) (*alexa.Response, error) {
	log.Println("Turn off!")
	return f.respBuilder.BasicResponse(req, alexa.ContextProperty{
		Namespace:                 alexa.NamespacePowerController,
		Name:                      "powerState",
		Value:                     json.RawMessage(`"` + "OFF" + `"`),
		TimeOfSample:              time.Now(),
		UncertaintyInMilliseconds: 500,
	}), nil
}

type windowControl struct {
	respBuilder *alexa.ResponseBuilder
}

func (w *windowControl) SetPercentage(ctx context.Context, req *alexa.Request) (*alexa.Response, error) {
	var targetPct alexa.SetPercentagePayload
	if err := json.Unmarshal(req.Directive.Payload, &targetPct); err != nil {
		return nil, fmt.Errorf("windowControl.SetPercentage: invalid payload: %v", err)
	}
	fmt.Printf("SetPercentage: %d\n", targetPct.Percentage)

	return w.respBuilder.BasicResponse(req, alexa.ContextProperty{
		Namespace:                 alexa.NamespacePercentageController,
		Name:                      "percentage",
		Value:                     w.marshalValue(targetPct.Percentage),
		TimeOfSample:              time.Now(),
		UncertaintyInMilliseconds: 500,
	}), nil
}

func (w *windowControl) AdjustPercentage(ctx context.Context, req *alexa.Request) (*alexa.Response, error) {
	var adjustPct alexa.AdjustPercentagePayload
	if err := json.Unmarshal(req.Directive.Payload, &adjustPct); err != nil {
		return nil, fmt.Errorf("windowControl.AdjustPercentage: invalid payload: %v", err)
	}
	fmt.Printf("AdjustPercentage: %d\n", adjustPct.PercentageDelta)

	return w.respBuilder.BasicResponse(req, alexa.ContextProperty{
		Namespace:                 alexa.NamespacePercentageController,
		Name:                      "percentage",
		Value:                     w.marshalValue(50),
		TimeOfSample:              time.Now(),
		UncertaintyInMilliseconds: 500,
	}), nil
}

func (w *windowControl) marshalValue(val uint8) json.RawMessage {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}

	return jsonVal
}
