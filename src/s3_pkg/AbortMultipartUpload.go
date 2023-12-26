// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// ///////////////////////////////////////
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

func (pkg *S3Pkg) AbortMultipartUpload(
	bucketName, objectName string, urlQuery url.Values,
) (resMp map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Error("s3.AbortMultipartUpload()", zap.Any("error", err))
		return nil, err
	}

	s3AbortMultipartUploadInput := &s3.AbortMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	if GetQueryString(urlQuery, "uploadId") != "" {
		uploadId := GetQueryString(urlQuery, "uploadId")
		s3AbortMultipartUploadInput.UploadId = &uploadId
	}

	Logger.Info("s3.AbortMultipartUpload()",
		zap.String("endpoint", pkg.Endpoint),
		zap.Any("Credentials", pkg.AwsConfig.Credentials),
		zap.Any("s3AbortMultipartUploadInput", s3AbortMultipartUploadInput))
	s3AbortRes, err := pkg.S3Client.AbortMultipartUpload(context.TODO(),
		s3AbortMultipartUploadInput)
	if err != nil ||
		(s3AbortRes.RequestCharged != "" && s3AbortRes.RequestCharged != "requester") {
		Logger.Error("s3.AbortMultipartUpload()", zap.Any("error", err))
		// HandleError(err)
		return nil, err
	}

	resMp = make(map[string]string)
	resMp["x-amz-request-charged"] = string(s3AbortRes.RequestCharged)

	respGetRawResponse := middleware.GetRawResponse(s3AbortRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*http.Response)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	Logger.Info("s3.AbortMultipartUpload(): ok.",
		zap.Any("result", resMp))

	return resMp, nil
}
