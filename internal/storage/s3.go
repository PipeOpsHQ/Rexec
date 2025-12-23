package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Store handles S3 storage operations for recordings
type S3Store struct {
	client   *s3.Client
	bucket   string
	prefix   string
	endpoint string
	region   string
}

// S3Config holds S3 configuration
type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        string // For S3-compatible services (MinIO, DigitalOcean Spaces, etc.)
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string // Optional prefix for all keys (e.g., "recordings/")
	ForcePathStyle  bool   // Use path-style addressing (required for some S3-compatible services)
}

// NewS3Store creates a new S3 store
func NewS3Store(cfg S3Config) (*S3Store, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Build AWS config options
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(cfg.Region))

	// Use static credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Build S3 client options
	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.ForcePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)

	store := &S3Store{
		client:   client,
		bucket:   cfg.Bucket,
		prefix:   cfg.Prefix,
		endpoint: cfg.Endpoint,
		region:   cfg.Region,
	}

	// Verify bucket access
	if err := store.verifyBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify S3 bucket access: %w", err)
	}

	log.Printf("[S3] Connected to bucket: %s (region: %s, prefix: %s)", cfg.Bucket, cfg.Region, cfg.Prefix)
	return store, nil
}

// verifyBucket checks if we can access the bucket
func (s *S3Store) verifyBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	return err
}

// key generates the S3 key for a recording
func (s *S3Store) key(recordingID string) string {
	return s.prefix + recordingID + ".cast"
}

// PutRecording uploads recording data to S3
func (s *S3Store) PutRecording(ctx context.Context, recordingID string, data []byte) error {
	key := s.key(recordingID)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/x-asciicast"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload recording to S3: %w", err)
	}

	log.Printf("[S3] Uploaded recording %s (%d bytes)", recordingID, len(data))
	return nil
}

// GetRecording downloads recording data from S3
func (s *S3Store) GetRecording(ctx context.Context, recordingID string) ([]byte, error) {
	key := s.key(recordingID)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recording from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read recording data: %w", err)
	}

	return data, nil
}

// DeleteRecording deletes recording data from S3
func (s *S3Store) DeleteRecording(ctx context.Context, recordingID string) error {
	key := s.key(recordingID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete recording from S3: %w", err)
	}

	log.Printf("[S3] Deleted recording %s", recordingID)
	return nil
}

// GetPresignedURL generates a presigned URL for downloading a recording
func (s *S3Store) GetPresignedURL(ctx context.Context, recordingID string, expiry time.Duration) (string, error) {
	key := s.key(recordingID)

	presignClient := s3.NewPresignClient(s.client)

	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:              aws.String(s.bucket),
		Key:                 aws.String(key),
		ResponseContentType: aws.String("application/x-asciicast"),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// RecordingExists checks if a recording exists in S3
func (s *S3Store) RecordingExists(ctx context.Context, recordingID string) (bool, error) {
	key := s.key(recordingID)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		var respErr interface{ HTTPStatusCode() int }
		if ok := err != nil; ok {
			if httpErr, ok := err.(interface{ HTTPStatusCode() int }); ok && httpErr.HTTPStatusCode() == http.StatusNotFound {
				return false, nil
			}
		}
		_ = respErr // silence unused warning
		return false, err
	}

	return true, nil
}

// PutFile uploads a generic file to S3
func (s *S3Store) PutFile(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	log.Printf("[S3] Uploaded file %s (%d bytes)", key, len(data))
	return nil
}

// GetFile downloads a generic file from S3
func (s *S3Store) GetFile(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	return data, nil
}

// Close closes the S3 store (no-op for S3)
func (s *S3Store) Close() error {
	return nil
}
