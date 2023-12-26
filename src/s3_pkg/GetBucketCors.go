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

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type CORSConfiguration struct {
	CORSRule []CORSRule
	// ResultMetadata middleware.Metadata
	// noSmithyDocumentSerde
}

type CORSRule struct {
	AllowedMethod []string
	AllowedOrigin []string
	AllowedHeader []string
	ExposeHeader  []string
	ID            *string
	MaxAgeSeconds int32
	// noSmithyDocumentSerde
}

func (pkg *S3Pkg) GetBucketCors(
	bucketName string, urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("GetBucketCors", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetBucketCorsInput := &s3.GetBucketCorsInput{
		Bucket: aws.String(bucketName),
	}

	s3GoaRes, err := pkg.S3Client.GetBucketCors(context.TODO(), s3GetBucketCorsInput)
	if err != nil {
		log.Println("[Error]:[S3Client.GetBucketCors]", err, '\n')
		// HandleError(err)
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3GoaRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	corsRule := CORSRule{}
	corsRules := make([]CORSRule, 0)
	for _, value := range s3GoaRes.CORSRules {
		corsRule.AllowedMethod = value.AllowedMethods
		corsRule.AllowedOrigin = value.AllowedOrigins
		corsRule.AllowedHeader = value.AllowedHeaders
		corsRule.ExposeHeader = value.ExposeHeaders
		corsRule.ID = value.ID
		corsRule.MaxAgeSeconds = value.MaxAgeSeconds
		corsRules = append(corsRules, corsRule)
	}
	goaRes := CORSConfiguration{CORSRule: corsRules}
	log.Println("GetBucketCors result: ", goaRes)

	resXML, err := xml.MarshalIndent(goaRes, " ", " ")
	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	log.Println("GetBucketCors resXML string():\n", string(res))

	return res, resMp, nil
}
