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
	"fmt"
	"net/url"
)

func (pkg *S3Pkg) GetBucketLogging(
	bucketName string, urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("GetBucketLogging", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetBucketLoggingInput := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}

	s3GblRes, err := pkg.S3Client.GetBucketLogging(context.TODO(), s3GetBucketLoggingInput)

	if err != nil {
		log.Println("[Error]:[S3Client.GetBucketLogging]", err, '\n')
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3GblRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	BucketLoggingStatus := s3GblRes
	log.Println("GetBucketLogging result: ", BucketLoggingStatus)

	resXML, err := xml.MarshalIndent(BucketLoggingStatus, " ", " ")

	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d",len(res))
	log.Println("GetBucketLogging resXML string():\n", string(res))

	return res, resMp, nil
}
