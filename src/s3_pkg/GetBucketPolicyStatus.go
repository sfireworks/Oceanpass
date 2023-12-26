package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
	"log"
	"net/url"
	"fmt"
)

func (pkg *S3Pkg) GetBucketPolicyStatus(
	bucketName string, urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("GetBucketPolicyStatus", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetBucketPolicyStatusInput := &s3.GetBucketPolicyStatusInput{
		Bucket: aws.String(bucketName),
	}

	s3GbpsRes, err := pkg.S3Client.GetBucketPolicyStatus(context.TODO(), s3GetBucketPolicyStatusInput)

	if err != nil {
		log.Println("[Error]:[S3Client.GetBucketPolicyStatus]", err, '\n')
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3GbpsRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	PolicyStatus := s3GbpsRes.PolicyStatus
	resXML, err := xml.MarshalIndent(PolicyStatus, " ", " ")
	if err != nil {
		log.Printf("marshal xml err: %v\n", err)
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d",len(res))
	log.Println("GetBucketPolicyStatus resXMl string():\n", string(res))

	return res, resMp, nil
}
