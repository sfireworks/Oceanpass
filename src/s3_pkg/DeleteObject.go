// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Zijun Hu
// ///////////////////////////////////////
package s3_pkg

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) DeleteObject(
	bucketName, objectName string,
	urlQuery url.Values,
) (resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.DeleteObject", zap.Any("error", err))
		return nil, err
	}

	s3DelObjInput := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	if versionid := GetQueryString(urlQuery, "versionId"); versionid != "" {
		s3DelObjInput.VersionId = &versionid
	}

	if bypassGovernanceRetention := GetQueryString(urlQuery, "x-amz-bypass-governance-retention"); bypassGovernanceRetention != "" {
		boolBGR, _ := strconv.ParseBool(bypassGovernanceRetention)
		s3DelObjInput.BypassGovernanceRetention = boolBGR
	}

	if expectedBucketOwner := GetQueryString(urlQuery, "x-amz-expected-bucket-owner"); expectedBucketOwner != "" {
		s3DelObjInput.ExpectedBucketOwner = &expectedBucketOwner
	}

	if mfa := GetQueryString(urlQuery, "x-amz-mfa"); mfa != "" {
		s3DelObjInput.MFA = &mfa
	}

	if requestPayer := GetQueryString(urlQuery, "x-amz-request-payer"); requestPayer != "" {
		s3DelObjInput.RequestPayer = types.RequestPayer(requestPayer)
	}

	result, err := s3pkg.S3Client.DeleteObject(context.TODO(), s3DelObjInput)
	if err != nil {
		Logger.Error("S3Pkg.DeleteObject",
			zap.Any("bucketName", bucketName), zap.Any("error", err))
		// HandleError(err)
		return nil, err
	}

	resMp = make(map[string]string)
	resMp["X-Amz-Delete-Marker"] = strconv.FormatBool(result.DeleteMarker)
	resMp["X-Amz-Version-Id"] = fmt.Sprintf("%p", result.VersionId)
	resMp["X-Amz-Request-Charged"] = string(result.RequestCharged)

	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*http.Response)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	Logger.Info("S3Pkg.DeleteObject(): ok.", zap.Any("result", resMp))

	return resMp, nil
}
