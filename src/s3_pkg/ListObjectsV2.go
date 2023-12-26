// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

// TODO: finish ListObjectsV2 about input handle
// paramsMap -> ListObjectsV2Input
func (pkg *S3Pkg) ListObjectsV2(
	bucketName string, urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("ListObjectsV2", zap.Any("error", err))
		return nil, nil, err
	}
	s3LoV2Input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	if GetQueryString(urlQuery, "max-keys") != "" {
		intMaxkeys, _ := strconv.Atoi(GetQueryString(urlQuery, "max-keys"))
		if intMaxkeys < 1 || intMaxkeys > 1000 {
			intMaxkeys = 1000
		}
		s3LoV2Input.MaxKeys = int32(intMaxkeys)
	} else {
		s3LoV2Input.MaxKeys = int32(100)
	}

	if GetQueryString(urlQuery, "prefix") != "" {
		prefix := GetQueryString(urlQuery, "prefix")
		s3LoV2Input.Prefix = &prefix
	}
	if GetQueryString(urlQuery, "continuation-token") != "" {
		continuationToken := GetQueryString(urlQuery, "continuation-token")
		s3LoV2Input.ContinuationToken = &continuationToken
	}
	if GetQueryString(urlQuery, "fetch-owner") != "" {
		fetchOwner := GetQueryString(urlQuery, "fetch-owner")
		boolFetchOwner, _ := strconv.ParseBool(fetchOwner)
		s3LoV2Input.FetchOwner = boolFetchOwner
	}
	if GetQueryString(urlQuery, "start-after") != "" {
		startAfter := GetQueryString(urlQuery, "start-after")
		s3LoV2Input.StartAfter = &startAfter
	}
	if GetQueryString(urlQuery, "delimiter") != "" {
		delimiter := GetQueryString(urlQuery, "delimiter")
		s3LoV2Input.Delimiter = &delimiter
	}
	if GetQueryString(urlQuery, "encoding-type") != "" {
		encodingType := GetQueryString(urlQuery, "encoding-type")
		s3LoV2Input.EncodingType = types.EncodingType(encodingType)
	}

	Logger.Info("ListObjectsV2()",
		zap.String("endpoint", pkg.Endpoint),
		zap.Any("Credentials", pkg.AwsConfig.Credentials),
		zap.Any("s3LoV2Input", s3LoV2Input))

	s3Lov2Res, err := pkg.S3Client.ListObjectsV2(context.TODO(), s3LoV2Input)
	if err != nil {
		Logger.Error("S3Pkg.ListObjectsV2 ",
			zap.Any("bucketName", bucketName),
			zap.Any("error", err))
		// HandleError(err)
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3Lov2Res.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	// log.Println("ListObjectsV2 result: ", lov2Res)

	lov2Res := EnumerationResult{
		CommonPrefixes:        s3Lov2Res.CommonPrefixes,
		Contents:              s3Lov2Res.Contents,
		ContinuationToken:     s3Lov2Res.ContinuationToken,
		Delimiter:             s3Lov2Res.Delimiter,
		EncodingType:          s3Lov2Res.EncodingType,
		IsTruncated:           s3Lov2Res.IsTruncated,
		KeyCount:              s3Lov2Res.KeyCount,
		MaxKeys:               s3Lov2Res.MaxKeys,
		Name:                  s3Lov2Res.Name,
		NextContinuationToken: s3Lov2Res.NextContinuationToken,
		Prefix:                s3Lov2Res.Prefix,
		StartAfter:            s3Lov2Res.StartAfter,
		ResultMetadata:        s3Lov2Res.ResultMetadata,
	}

	resXML, err := xml.MarshalIndent(lov2Res, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.ListObjectsV2(): marshal xml err ",
			zap.Any("error", err))
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))

	Logger.Info("S3Pkg.ListObjectsV2(): ok.",
		zap.Any("result", string(res)))

	return res, resMp, nil
}
