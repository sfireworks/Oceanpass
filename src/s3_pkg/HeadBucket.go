package s3_pkg

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) HeadBucket(
	bucketName string, urlQuery url.Values) (resMp map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.HeadBucket: call GetS3Client() ", zap.Any("error", err))
		return nil, err
	}

	s3HeadBucketInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	Logger.Info("S3Pkg.HeadBucket(): ", zap.Any("s3HeadBucketInput", s3HeadBucketInput))
	result, err := pkg.S3Client.HeadBucket(context.TODO(), s3HeadBucketInput)
	if err != nil {
		Logger.Error("S3Pkg.HeadBucket", zap.Any("bucketName", bucketName),
			zap.Any("error", err))
		//default:404 not found
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*http.Response)

	resMp = make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("S3Pkg.HeadBucket(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
