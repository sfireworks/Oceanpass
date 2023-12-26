// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// Author: Zijun Hu
// ///////////////////////////////////////
package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/url"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) CompleteMultipartUpload(
	bucketName, objectName string, urlQuery url.Values,
	partNumEtag map[int]string,
) (res []byte, resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.CompleteMultipartUpload", zap.Any("error", err))
		return nil, resMp, err
	}

	// sort PartNum
	SortedPartkeys := make([]int, 0, len(partNumEtag))
	for k, _ := range partNumEtag {
		SortedPartkeys = append(SortedPartkeys, k)
	}
	sort.Ints(SortedPartkeys)

	parts := types.CompletedMultipartUpload{}
	for _, partNum := range SortedPartkeys {
		strPartEtg := partNumEtag[partNum]
		parts.Parts = append(parts.Parts,
			types.CompletedPart{
				ETag:       &strPartEtg,
				PartNumber: int32(partNum),
			})

	}

	s3CompleteMPUpInput := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucketName),
		Key:             aws.String(objectName),
		UploadId:        aws.String(urlQuery["uploadId"][0]),
		MultipartUpload: &parts,
	}

	if requestPayer := GetQueryString(urlQuery, "x-amz-request-payer"); requestPayer != "" {
		s3CompleteMPUpInput.RequestPayer = types.RequestPayer(requestPayer)
	}

	s3CompleteMPUpOutput, err := s3pkg.S3Client.CompleteMultipartUpload(context.TODO(), s3CompleteMPUpInput)
	if err != nil {
		Logger.Error("S3Pkg.CompleteMultipartUpload ",
			zap.Any("bucketName", bucketName),
			zap.Any("objectName", objectName),
			zap.Any("error", err))
		return nil, resMp, err
	}
	respGetRawResponse := middleware.GetRawResponse(s3CompleteMPUpOutput.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp = make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	resXML, err := xml.MarshalIndent(s3CompleteMPUpOutput, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.CompleteMultipartUpload(): marshal xml err ",
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

	Logger.Info("S3Pkg.CompleteMultipartUpload(): ok.",
		zap.Any("result", string(res)))

	return res, resMp, nil
}
