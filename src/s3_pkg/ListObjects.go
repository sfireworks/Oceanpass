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

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func getObjectsFormResponse(lores oss.ListObjectsResult) string {
	var output string
	for _, object := range lores.Objects {
		output += object.Key + "  "
	}
	return output
}

type ListObjectsReq struct {
	Delimiter string `json:"delimiter"`
	Marker    string `json:"marker"`
	Maxkeys   int64  `json:"max-keys"`
	Prefix    string `json:"prefix"`
}

// check the paramters' validity
func CheckListObjectsParams(params ListObjectsReq) (err error) {
	//log.Println(params)
	return
}

func (pkg *S3Pkg) GetHttpQueryParams(vars url.Values) (map[string]string, bool) {
	mp := make(map[string]string)
	mp["delimiter"] = GetQueryString(vars, "delimiter")
	mp["marker"] = GetQueryString(vars, "marker")
	mp["max-keys"] = "1000"
	if GetQueryString(vars, "max-keys") != "" {
		max_keys := GetQueryString(vars, "max-keys")
		intMaxkeys, _ := strconv.Atoi(max_keys)
		if intMaxkeys >= 2 || intMaxkeys <= 1000 {
			mp["max-keys"] = max_keys
		}
	}
	mp["prefix"] = GetQueryString(vars, "prefix")

	flag := false
	for k, _ := range mp {
		if mp[k] != "" {
			flag = true
		}
	}
	return mp, flag
}

// TODO: finish it and change name to ListObjects
func (pkg *S3Pkg) ListObjects(
	bucketName string,
	urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("ListObjects", zap.Any("error", err))
		return nil, nil, err
	}
	s3LoInput := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}

	if GetQueryString(urlQuery, "delimiter") != "" {
		delimiter := GetQueryString(urlQuery, "delimiter")
		s3LoInput.Delimiter = &delimiter
	}
	if GetQueryString(urlQuery, "marker") != "" {
		marker := GetQueryString(urlQuery, "marker")
		s3LoInput.Marker = &marker
	}
	if GetQueryString(urlQuery, "max-keys") != "" {
		intMaxkeys, _ := strconv.Atoi(GetQueryString(urlQuery, "max-keys"))
		if intMaxkeys < 1 || intMaxkeys > 1000 {
			intMaxkeys = 1000
		}
		s3LoInput.MaxKeys = int32(intMaxkeys)
	} else {
		s3LoInput.MaxKeys = int32(100)
	}
	if GetQueryString(urlQuery, "prefix") != "" {
		prefix := GetQueryString(urlQuery, "prefix")
		s3LoInput.Prefix = &prefix
	}

	// TODO: check this param
	// if GetQueryString(urlQuery, "encoding-type") != "" {
	// 	encodingType := GetQueryString(urlQuery, "encoding-type")
	// }
	s3LoInput.EncodingType = types.EncodingTypeUrl

	// TODO: finish this param in Header
	// if GetQueryString(xxx, "ExpectedBucketOwner") != "" {
	// 	expectedBucketOwner := GetQueryString(xxx, "ExpectedBucketOwner")
	// 	s3LoInput.ExpectedBucketOwner = &expectedBucketOwner
	// }

	Logger.Info("ListObjects()",
		zap.String("endpoint", pkg.Endpoint),
		zap.Any("Credentials", pkg.AwsConfig.Credentials),
		zap.Any("s3LoInput", s3LoInput))

	s3LoRes, err := pkg.S3Client.ListObjects(context.TODO(), s3LoInput)

	if err != nil {
		Logger.Error("S3Pkg.ListObjects ",
			zap.Any("bucketName", bucketName),
			zap.Any("error", err))
		HandleError(err)
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3LoRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	loRes := EnumerationResult{
		CommonPrefixes: s3LoRes.CommonPrefixes,
		Contents:       s3LoRes.Contents,
		Delimiter:      s3LoRes.Delimiter,
		EncodingType:   s3LoRes.EncodingType,
		IsTruncated:    s3LoRes.IsTruncated,
		MaxKeys:        s3LoRes.MaxKeys,
		Name:           s3LoRes.Name,
		Prefix:         s3LoRes.Prefix,
		ResultMetadata: s3LoRes.ResultMetadata,
		Marker:         s3LoRes.Marker,
		NextMarker:     s3LoRes.NextMarker,
	}

	resXML, err := xml.MarshalIndent(loRes, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.ListObjects(): marshal xml err ",
			zap.Any("error", err))
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	return res, resMp, nil
}
