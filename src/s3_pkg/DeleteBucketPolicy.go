package s3_pkg

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) DeleteBucketPolicy(bucketName string) error {
	err := s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.DeleteBucketPolicy: call GetS3Client() ", zap.Any("error", err))
		return err
	}
	s3DeleteBucketPolicyInput := &s3.DeleteBucketPolicyInput{Bucket: aws.String(bucketName)}
	result, err := s3pkg.S3Client.DeleteBucketPolicy(context.TODO(), s3DeleteBucketPolicyInput)
	if err != nil {
		Logger.Error("S3Pkg.DeleteBucketPolicy", zap.Any("bucketName", bucketName), zap.Any("error", err))
		return err
	}
	Logger.Info("S3Pkg.DeleteBucketPolicy(): ok.", zap.Any("result", result))
	return nil
}
