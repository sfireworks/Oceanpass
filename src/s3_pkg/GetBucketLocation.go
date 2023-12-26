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

func (pkg *S3Pkg) GetBucketLocation(
	bucketName string, urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("GetBucketLocation", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetBucketLocationInput := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}

	s3GblRes, err := pkg.S3Client.GetBucketLocation(context.TODO(), s3GetBucketLocationInput)
	if err != nil {
		log.Println("[Error]:[S3Client.GetBucketLocation]", err, '\n')
		// HandleError(err)
		return nil, nil, err
	}
	// log.Println("GetBucketLocation result: ", s3GblRes)
	respGetRawResponse := mid.GetRawResponse(s3GblRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	LocationConstraint := s3GblRes.LocationConstraint
	resXML, err := xml.MarshalIndent(LocationConstraint, " ", " ")
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
	log.Println("GetBucketLocation resXML string():\n", string(res))

	return res, resMp, nil
}
