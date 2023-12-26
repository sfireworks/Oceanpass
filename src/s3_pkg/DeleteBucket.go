package s3_pkg

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) DeleteBucket(bucketName string) (resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.DeleteBucket", zap.Any("error", err))
		return nil, err
	}
	s3DelBucInput := &s3.DeleteBucketInput{Bucket: aws.String(bucketName)}
	s3DelBucRes, err := s3pkg.S3Client.DeleteBucket(context.TODO(), s3DelBucInput)
	if err != nil {
		Logger.Error("S3Pkg.DeleteBucket",
			zap.Any("bucketName", bucketName), zap.Any("error", err))
		return nil, err
	}

	resMp = make(map[string]string)
	respGetRawResponse := middleware.GetRawResponse(s3DelBucRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*http.Response)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	Logger.Info("S3Pkg.DeleteBucket(): ok.", zap.Any("result", resMp))

	return resMp, nil
}
