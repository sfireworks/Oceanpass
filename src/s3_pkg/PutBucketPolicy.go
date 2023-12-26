package s3_pkg

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"

	"go.uber.org/zap"
)

func (s3pkg S3Pkg) PutBucketPolicy(bucketName string, r *http.Request) (map[string]string, error) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	policyInfo := string(bodyBytes)
	err := s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.PutBucketPolicy", zap.Any("error", err))
		return nil, err
	}
	s3PutBucketPolicyInput := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policyInfo),
	}
	Logger.Info("S3Pkg.PutBucketPolicy()",
		zap.String("endpoint", s3pkg.Endpoint),
		zap.Any("Credentials", s3pkg.AwsConfig.Credentials),
		zap.Any("policyInfo", policyInfo),
		zap.Any("s3LoV2Input", s3PutBucketPolicyInput))
	result, err := s3pkg.S3Client.PutBucketPolicy(context.TODO(), s3PutBucketPolicyInput)
	if err != nil {
		Logger.Error("S3Pkg.PutBucketPolicy", zap.Any("error", err))
		// HandleError(err)
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("S3Pkg.PutBucketPolicy(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
