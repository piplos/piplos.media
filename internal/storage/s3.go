package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Config describes a connection to an S3-compatible service.
// For Cloudflare R2 the endpoint is https://<account_id>.r2.cloudflarestorage.com.
type S3Config struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	// UsePathStyle enables path-style addressing (MinIO and some providers).
	UsePathStyle bool `json:"use_path_style"`
}

// Validate checks required connection fields.
func (c *S3Config) Validate() error {
	if strings.TrimSpace(c.Endpoint) == "" {
		return errors.New("S3 endpoint is required")
	}
	if strings.TrimSpace(c.Bucket) == "" {
		return errors.New("S3 bucket is required")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.SecretAccessKey) == "" {
		return errors.New("S3 credentials are required")
	}
	return nil
}

// S3 is a Storage backed by an S3-compatible bucket.
type S3 struct {
	client *s3.Client
	bucket string
}

// NewS3 creates an S3 storage client. Checksum calculation is relaxed to
// "when required" for Cloudflare R2 compatibility (R2 rejects default CRC32).
func NewS3(cfg S3Config) (*S3, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	region := cfg.Region
	if region == "" {
		region = "auto"
	}
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	client := s3.New(s3.Options{
		Region:                     region,
		BaseEndpoint:               aws.String(endpoint),
		Credentials:                credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle:               cfg.UsePathStyle,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &S3{client: client, bucket: cfg.Bucket}, nil
}

// Put uploads an object with a known size.
func (s *S3) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return fmt.Errorf("s3 put %s: %w", key, err)
	}
	return nil
}

// Get downloads an object.
func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, 0, fmt.Errorf("s3 get %s: %w", key, fs.ErrNotExist)
		}
		return nil, 0, fmt.Errorf("s3 get %s: %w", key, err)
	}
	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, size, nil
}

// List returns all objects under prefix (paginated internally).
func (s *S3) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	out := []ObjectInfo{}
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 list %s: %w", prefix, err)
		}
		for _, obj := range page.Contents {
			info := ObjectInfo{Key: aws.ToString(obj.Key)}
			if obj.Size != nil {
				info.Size = *obj.Size
			}
			if obj.LastModified != nil {
				info.ModTime = *obj.LastModified
			}
			out = append(out, info)
		}
	}
	return out, nil
}

// Delete removes an object (S3 delete of a missing key succeeds).
func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %s: %w", key, err)
	}
	return nil
}

// TestConnection verifies credentials and bucket access.
func (s *S3) TestConnection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("s3 connection failed: %w", err)
	}
	return nil
}
