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
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type ListPartsOutput struct {
	AbortDate            *time.Time
	AbortRuleId          *string
	Bucket               *string
	ChecksumAlgorithm    types.ChecksumAlgorithm
	Initiator            *types.Initiator
	IsTruncated          bool
	Key                  *string
	MaxParts             int32
	NextPartNumberMarker *string
	Owner                *types.Owner
	PartNumberMarker     *string
	Part                 []types.Part
	RequestCharged       types.RequestCharged
	StorageClass         types.StorageClass
	UploadId             *string
	ResultMetadata       middleware.Metadata
}

func (pkg *S3Pkg) ListParts(
	bucketName, objectName string,
	urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("ListParts", zap.Any("error", err))
		return nil, nil, err
	}

	s3ListPartsInput := &s3.ListPartsInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	if GetQueryString(urlQuery, "uploadId") != "" {
		uploadId := GetQueryString(urlQuery, "uploadId")
		s3ListPartsInput.UploadId = &uploadId
	}
	if GetQueryString(urlQuery, "max-parts") != "" {
		maxParts := GetQueryString(urlQuery, "max-parts")
		iMaxParts, _ := strconv.Atoi(maxParts)
		i32MaxParts := int32(iMaxParts)
		s3ListPartsInput.MaxParts = i32MaxParts
	}
	if GetQueryString(urlQuery, "part-number-marker") != "" {
		partNumberMarker := GetQueryString(urlQuery, "part-number-marker")
		s3ListPartsInput.PartNumberMarker = &partNumberMarker
	}

	s3LpRes, err := pkg.S3Client.ListParts(context.TODO(), s3ListPartsInput)
	if err != nil {
		log.Println("[Error]:[S3Client.ListParts]", err, '\n')
		log.Println("[Error]:[S3Client.ListParts] ObjectKey", objectName)
		// HandleError(err)
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3LpRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	lpRes := ListPartsOutput{
		AbortDate:            s3LpRes.AbortDate,
		AbortRuleId:          s3LpRes.AbortRuleId,
		Bucket:               s3LpRes.Bucket,
		ChecksumAlgorithm:    s3LpRes.ChecksumAlgorithm,
		Initiator:            s3LpRes.Initiator,
		IsTruncated:          s3LpRes.IsTruncated,
		Key:                  s3LpRes.Key,
		MaxParts:             s3LpRes.MaxParts,
		NextPartNumberMarker: s3LpRes.NextPartNumberMarker,
		Owner:                s3LpRes.Owner,
		PartNumberMarker:     s3LpRes.PartNumberMarker,
		Part:                 s3LpRes.Parts,
		RequestCharged:       s3LpRes.RequestCharged,
		StorageClass:         s3LpRes.StorageClass,
		UploadId:             s3LpRes.UploadId,
		ResultMetadata:       s3LpRes.ResultMetadata,
	}
	// log.Println("ListParts result: ", lpRes)

	resXML, err := xml.MarshalIndent(lpRes, " ", " ")
	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, nil, err
	}

	// Add Header and "\n\r"
	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	log.Println("ListParts resXML string():\n", string(res))

	return res, resMp, nil
}
