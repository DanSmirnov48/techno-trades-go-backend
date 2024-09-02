package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Client is a wrapper around the S3 service client
type S3Client struct {
	s3         *s3.S3
	bucketName string
}

// NewS3Client creates a new S3 client
func NewS3Client() (*S3Client, error) {
	// Initialize an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)
	return &S3Client{
		s3:         s3Client,
		bucketName: os.Getenv("AWS_S3_BUCKET_NAME"),
	}, nil
}

// UploadFile uploads a file to S3 and returns the URL of the uploaded file
func (s *S3Client) UploadFile(file *multipart.FileHeader) (string, error) {
	// Open the file
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	// Read file content into a byte array
	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Generate the file key with the avatars folder
	fileKey := fmt.Sprintf("avatars/%s", filepath.Base(file.Filename))

	// Upload the file to S3
	_, err = s.s3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.bucketName),
		Key:                  aws.String(fileKey),
		Body:                 bytes.NewReader(fileBytes),
		ACL:                  aws.String("public-read"), // Make the file publicly accessible
		ContentLength:        aws.Int64(file.Size),
		ContentType:          aws.String(file.Header.Get("Content-Type")),
		ContentDisposition:   aws.String("inline"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, fileKey)
	return fileURL, nil
}

// GetFile retrieves a file from S3 and returns its content
func (s *S3Client) GetFile(fileKey string) ([]byte, error) {
	// Retrieve the file from S3
	result, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file from S3: %v", err)
	}
	defer result.Body.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	return fileBytes, nil
}