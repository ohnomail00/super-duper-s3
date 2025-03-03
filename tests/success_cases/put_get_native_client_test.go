package success_cases

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ohnomail00/super-duper-s3/tests/utils"
)

func TestS3UploadDownload(t *testing.T) {
	storage, _ := utils.StartTestStorageServer(t)
	gateway, _ := utils.StartTestGatewayServer(t, []string{storage.URL})
	awsURL := gateway.URL

	t.Logf("Using S3 endpoint: %s", awsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:           awsURL,
						SigningRegion: "us-east-1",
					}, nil
				}
				return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
			},
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	bucket := "mybucket"
	key := "testobject.txt"
	content := "Hello, S3 test from our custom server!"
	body := bytes.NewReader([]byte(content))

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   body,
	})
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	getResp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		t.Fatalf("failed to get object: %v", err)
	}
	downloaded, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("failed to read object body: %v", err)
	}
	getResp.Body.Close()

	if string(downloaded) != content {
		t.Fatalf("object content mismatch: expected %q, got %q", content, string(downloaded))
	}

}
