package s3store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"golang.org/x/oauth2"
)

// TokenStorage uses S3 as a simple backing store for a user's oauth tokens.
// Tokens are stored as json documents named by the user's id.
// This isn't the most secure option although it can be improved by enabling
// encryption and strictly limiting access to the S3 bucket.
// Due to S3's eventually consistent nature a Read may not always reflect the
// lastest tokens provided to Write.
type TokenStorage struct {
	S3     s3iface.S3API
	Bucket string
}

func (s *TokenStorage) Write(ctx context.Context, id string, token *oauth2.Token) error {
	content, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %v", err)
	}

	req := s3.PutObjectInput{
		Bucket:      &s.Bucket,
		Key:         &id,
		Body:        bytes.NewReader(content),
		ContentType: aws.String("application/json"),
	}

	if _, err := s.S3.PutObjectWithContext(ctx, &req); err != nil {
		return fmt.Errorf("failed to upload to s3: %v", err)
	}

	return nil
}

func (s *TokenStorage) Read(ctx context.Context, id string) (*oauth2.Token, error) {
	req := s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &id,
	}

	resp, err := s.S3.GetObjectWithContext(ctx, &req)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("failed to retrieve from s3: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read s3 data: %v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	return &token, nil
}
