package adapter

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type ObjectStorage struct {
	Client *s3.S3
	Bucket string
}

func SetupObjectStorage(endpoint, accessKeyPath, secretKeyPath string, useSSL bool, region, bucket string) (*ObjectStorage, error) {
	accessKeyBytes, err := ioutil.ReadFile(accessKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading secret access key from %s: %v", accessKeyPath, err)
	}

	secretKeyBytes, err := ioutil.ReadFile(secretKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading secret access key from %s: %v", secretKeyPath, err)
	}

	accessKey := strings.TrimSpace(string(accessKeyBytes))
	secretKey := strings.TrimSpace(string(secretKeyBytes))

	endpointProtocol := "http"
	if useSSL {
		endpointProtocol = "https"
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(fmt.Sprintf("%s://%s", endpointProtocol, endpoint)),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(!useSSL),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("set up aws session: %v", err)
	}

	s3Client := s3.New(newSession)

	return &ObjectStorage{
		Client: s3Client,
		Bucket: bucket,
	}, nil
}

func (o *ObjectStorage) Presign(ctx context.Context, path string) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "s3: presign")
	defer span.Finish()

	span.LogFields(
		log.String("path", path),
	)

	req, _ := o.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(o.Bucket),
		Key:    aws.String(path),
	})

	url, err := req.Presign(10 * time.Minute)
	if err != nil {
		span.LogFields(log.Error(err))
		return "", fmt.Errorf("presign path %s: %v", path, err)
	}

	return url, nil
}

func (o *ObjectStorage) Upload(ctx context.Context, path string, file io.ReadSeeker, contentType string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "s3: upload")
	defer span.Finish()

	span.LogFields(
		log.String("path", path),
	)

	put := &s3.PutObjectInput{
		Bucket: aws.String(o.Bucket),
		Key:    aws.String(path),
		Body:   file,
	}

	if contentType != "" {
		put.ContentType = aws.String(contentType)
	}

	_, err := o.Client.PutObjectWithContext(ctx, put)
	if err != nil {
		span.LogFields(log.Error(err))
		return err
	}

	return nil
}

func (o *ObjectStorage) Exists(ctx context.Context, path string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "s3: exists")
	defer span.Finish()

	span.LogFields(
		log.String("path", path),
	)

	head := &s3.HeadObjectInput{
		Bucket: aws.String(o.Bucket),
		Key:    aws.String(path),
	}

	_, err := o.Client.HeadObjectWithContext(ctx, head)
	if err != nil {
		span.LogFields(log.Error(err))
		return err
	}

	return nil
}
