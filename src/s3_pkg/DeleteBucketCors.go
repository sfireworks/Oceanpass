package s3_pkg

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg S3Pkg) DeleteBucketCors(bucketName string) (resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.DeleteBucketCors::GetS3Client()", zap.Any("error", err))
		return nil, err
	}
	s3DeleteBucketCorsInput := &s3.DeleteBucketCorsInput{Bucket: aws.String(bucketName)}
	result, err := s3pkg.S3Client.DeleteBucketCors(context.TODO(), s3DeleteBucketCorsInput)
	if err != nil {
		Logger.Error("S3Pkg.DeleteBucketCors",
			zap.String("bucketName", bucketName),
			zap.Any("error", err))
		return nil, err
	}

	resMp = make(map[string]string)
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*http.Response)
	resMp = make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	Logger.Info("S3Pkg.DeleteObject(): ok.", zap.Any("result", resMp))

	return resMp, nil
}
