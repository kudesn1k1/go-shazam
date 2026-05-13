package containers

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

type MinioInfo struct {
	Endpoint  string // http://host:port
	AccessKey string
	SecretKey string
	Bucket    string
	Client    *s3.Client
}

// StartMinio boots an ephemeral MinIO container, creates a fresh bucket, and
// returns an S3 client wired to it.
func StartMinio(t *testing.T) MinioInfo {
	t.Helper()
	ctx := context.Background()

	access := "testminio"
	secret := "testminio-secret"
	bucket := "test-bucket"

	c, err := tcminio.Run(ctx, "minio/minio:latest",
		tcminio.WithUsername(access),
		tcminio.WithPassword(secret),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	hostPort, err := c.ConnectionString(ctx)
	require.NoError(t, err)
	endpoint := "http://" + hostPort

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(access, secret, "")),
	)
	require.NoError(t, err)

	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	_, err = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	require.NoError(t, err)

	return MinioInfo{
		Endpoint:  endpoint,
		AccessKey: access,
		SecretKey: secret,
		Bucket:    bucket,
		Client:    cli,
	}
}
