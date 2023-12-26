// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Zijun Hu
// ///////////////////////////////////////
package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) CreateMultipartUpload(
	bucketName string, objectName string, urlQuery url.Values, header http.Header,
) (res []byte, resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.CreateMultipartUpload", zap.Any("error", err))
		return nil, resMp, err
	}
	s3CMPUpInput := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}
	metaMp := make(map[string]string)
	for key, val := range header {
		if strings.HasPrefix(key, "X-Amz-Meta-") || strings.HasPrefix(key, "X-Oss-Meta-") {
			metaKey := key[len("X-Oss-Meta-"):]
			metaKey = strings.ToLower(metaKey)
			metaMp[metaKey] = val[0]
		}
	}
	s3CMPUpInput.Metadata = metaMp
	if requestPayer := GetQueryString(urlQuery, "x-amz-request-payer"); requestPayer != "" {
		s3CMPUpInput.RequestPayer = types.RequestPayer(requestPayer)
	}
	if header.Get("Content-Type") != "" {
		s3CMPUpInput.ContentType = aws.String(header.Get("Content-Type"))
	}
	if header.Get("X-Oss-Server-Side-Encryption") != "" {
		s3CMPUpInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Oss-Server-Side-Encryption"))
	}
	if header.Get("X-Amz-Server-Side-Encryption") != "" {
		s3CMPUpInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Amz-Server-Side-Encryption"))
	}
	s3CMPUpOutput, err := s3pkg.S3Client.CreateMultipartUpload(context.TODO(), s3CMPUpInput)
	if err != nil {
		Logger.Error("S3Pkg.CreateMultipartUpload ",
			zap.Any("objectName", objectName), zap.Any("error", err))
		return nil, resMp, err
	}
	respGetRawResponse := middleware.GetRawResponse(s3CMPUpOutput.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp = make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	resXML, err := xml.MarshalIndent(s3CMPUpOutput, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.CreateMultipartUpload(): marshal xml err ",
			zap.Any("error", err))
		return nil, resMp, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res = buffer.Bytes()

	finLen := buffer.Len()
	resMp["Content-Length"] = strconv.Itoa(finLen)

	Logger.Info("S3Pkg.CreateMultipartUpload(): ok.",
		zap.Any("result", string(res)))

	return res, resMp, nil
}
