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

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type ListMultipartUploadsOutput struct {
	Bucket             *string
	CommonPrefixes     []types.CommonPrefix
	Delimiter          *string
	EncodingType       types.EncodingType
	IsTruncated        bool
	KeyMarker          *string
	MaxUploads         int32
	NextKeyMarker      *string
	NextUploadIdMarker *string
	Prefix             *string
	UploadIdMarker     *string
	Upload             []types.MultipartUpload
	ResultMetadata     middleware.Metadata
	// noSmithyDocumentSerde
}

func (pkg *S3Pkg) ListMultipartUploads(
	bucketName, objectName string,
	urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("ListMultipartUploads", zap.Any("error", err))
		return nil, nil, err
	}

	s3ListMultipartUploadsInput := &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucketName),
	}

	if GetQueryString(urlQuery, "delimiter") != "" {
		delimiter := GetQueryString(urlQuery, "delimiter")
		s3ListMultipartUploadsInput.Delimiter = &delimiter
	}
	s3ListMultipartUploadsInput.EncodingType = types.EncodingTypeUrl

	if GetQueryString(urlQuery, "key-marker") != "" {
		keyMarker := GetQueryString(urlQuery, "key-marker")
		s3ListMultipartUploadsInput.KeyMarker = &keyMarker
	}
	if GetQueryString(urlQuery, "max-uploads") != "" {
		maxUploads := GetQueryString(urlQuery, "max-uploads")
		iMaxUploads, _ := strconv.Atoi(maxUploads)
		i32MaxUploads := int32(iMaxUploads)
		s3ListMultipartUploadsInput.MaxUploads = i32MaxUploads
	}
	if GetQueryString(urlQuery, "prefix") != "" {
		prefix := GetQueryString(urlQuery, "prefix")
		s3ListMultipartUploadsInput.Prefix = &prefix
	}
	if GetQueryString(urlQuery, "upload-id-marker") != "" {
		uploadIdMarker := GetQueryString(urlQuery, "upload-id-marker")
		s3ListMultipartUploadsInput.UploadIdMarker = &uploadIdMarker
	}

	S3LmpuRes, err := pkg.S3Client.ListMultipartUploads(context.TODO(), s3ListMultipartUploadsInput)
	if err != nil {
		log.Println("[Error]:[S3Client.ListMultipartUploads]", err, '\n')
		log.Println("[Error]:[S3Client.ListMultipartUploads] ObjectKey", objectName)
		// HandleError(err)
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(S3LmpuRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	lmpuRes := ListMultipartUploadsOutput{
		Bucket:             S3LmpuRes.Bucket,
		CommonPrefixes:     S3LmpuRes.CommonPrefixes,
		Delimiter:          S3LmpuRes.Delimiter,
		EncodingType:       S3LmpuRes.EncodingType,
		IsTruncated:        S3LmpuRes.IsTruncated,
		KeyMarker:          S3LmpuRes.KeyMarker,
		MaxUploads:         S3LmpuRes.MaxUploads,
		NextKeyMarker:      S3LmpuRes.NextKeyMarker,
		NextUploadIdMarker: S3LmpuRes.NextUploadIdMarker,
		Prefix:             S3LmpuRes.Prefix,
		UploadIdMarker:     S3LmpuRes.UploadIdMarker,
		Upload:             S3LmpuRes.Uploads,
		ResultMetadata:     S3LmpuRes.ResultMetadata,
	}

	log.Println("ListMultipartUploads result: ", lmpuRes)

	resXML, err := xml.MarshalIndent(lmpuRes, " ", " ")
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
	log.Println("ListMultipartUploads resXML string():\n", string(res))

	return res, resMp, nil
}
