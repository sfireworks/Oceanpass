package s3_pkg

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) CreateBucket(bucketName string) (map[string]string, error) {
	err := s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.CreateBucket", zap.Any("error", err))
		return nil, err
	}
	S3CBInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	Logger.Info("S3Pkg.CreateBucket()",
		zap.String("endpoint", s3pkg.Endpoint),
		zap.Any("Credentials", s3pkg.AwsConfig.Credentials),
		zap.Any("S3CBInput", S3CBInput))
	result, err := s3pkg.S3Client.CreateBucket(context.TODO(), S3CBInput)

	if err != nil {
		//default:404 not found
		Logger.Error("S3Pkg.CreateBucket(): ", zap.Any("error", err))
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("S3Pkg.CreateBucket(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
