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
func newS3Client() (*S3Client, error) {
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
func UploadFile(file *multipart.FileHeader) (string, error) {
	// Create a new S3 client
	s3Client, err := newS3Client()
	if err != nil {
		return "", err
	}

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
	fileKey := fmt.Sprintf("users/%s", filepath.Base(file.Filename))

	// Upload the file to S3
	_, err = s3Client.s3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s3Client.bucketName),
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
	fileURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Client.bucketName, fileKey)
	return fileURL, nil
}

// GetFile retrieves a file from S3 and returns its content
func GetFile(fileKey string) ([]byte, error) {
	// Create a new S3 client
	s3Client, err := newS3Client()
	if err != nil {
		return nil, err
	}
	// Retrieve the file from S3
	result, err := s3Client.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Client.bucketName),
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

// DeleteFile deletes a file from the S3 bucket
func DeleteFile(fileKey string) error {
	// Create a new S3 client
	s3Client, err := newS3Client()
	if err != nil {
		return err
	}

	// Create the delete input request
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3Client.bucketName),
		Key:    aws.String(fileKey),
	}

	// Perform the delete operation
	_, err = s3Client.s3.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %v", err)
	}

	// Wait until the file is deleted (optional but recommended)
	err = s3Client.s3.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s3Client.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to wait for file deletion from S3: %v", err)
	}

	return nil
}
