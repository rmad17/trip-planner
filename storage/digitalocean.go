package storage

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// DigitalOceanProvider implements StorageProvider for DigitalOcean Spaces
type DigitalOceanProvider struct {
	client   *s3.S3
	uploader *s3manager.Uploader
	bucket   string
	region   string
	baseURL  string
}

// DigitalOceanConfig holds DigitalOcean Spaces configuration
type DigitalOceanConfig struct {
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"`
	Region          string `json:"region" yaml:"region"`
	Bucket          string `json:"bucket" yaml:"bucket"`
	Endpoint        string `json:"endpoint" yaml:"endpoint"` // e.g., "nyc3.digitaloceanspaces.com"
}

// NewDigitalOceanProvider creates a new DigitalOcean Spaces storage provider
func NewDigitalOceanProvider(config DigitalOceanConfig) (*DigitalOceanProvider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = config.Region + ".digitaloceanspaces.com"
	}

	awsConfig := &aws.Config{
		Region:           aws.String(config.Region),
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Endpoint:         aws.String("https://" + endpoint),
		S3ForcePathStyle: aws.Bool(false),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	baseURL := "https://" + config.Bucket + "." + endpoint

	return &DigitalOceanProvider{
		client:   client,
		uploader: uploader,
		bucket:   config.Bucket,
		region:   config.Region,
		baseURL:  baseURL,
	}, nil
}

// Upload uploads a file to DigitalOcean Spaces
func (d *DigitalOceanProvider) Upload(key string, file io.Reader, contentType string) (*UploadResult, error) {
	result, err := d.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(d.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"), // Make files publicly readable
	})
	if err != nil {
		return nil, err
	}

	// Get object info for size
	headResult, err := d.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		Key:         key,
		URL:         result.Location,
		Size:        *headResult.ContentLength,
		ContentType: contentType,
		Provider:    "digitalocean",
	}, nil
}

// Download downloads a file from DigitalOcean Spaces
func (d *DigitalOceanProvider) Download(key string) (io.ReadCloser, error) {
	result, err := d.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// Delete removes a file from DigitalOcean Spaces
func (d *DigitalOceanProvider) Delete(key string) error {
	_, err := d.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetURL generates a presigned URL for accessing the file
func (d *DigitalOceanProvider) GetURL(key string, expiry time.Duration) (string, error) {
	req, _ := d.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiry)
	return url, err
}

// GetPublicURL returns a public URL for the file
func (d *DigitalOceanProvider) GetPublicURL(key string) string {
	return d.baseURL + "/" + key
}

// GetProvider returns the provider name
func (d *DigitalOceanProvider) GetProvider() string {
	return "digitalocean"
}