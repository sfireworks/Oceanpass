package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) UploadPartCopy(bucketName string, objectName string,
	partNumber, uploadId string, copySource string) (map[string]string, []byte, error) {
	err := pkg.GetS3Client()
	if err != nil {
		Logger.Info("UploadPartCopy", zap.Any("error", err))
		return nil, nil, err
	}
	partNumberInt, _ := strconv.ParseInt(partNumber, 10, 32)

	resp, err := pkg.S3Client.UploadPartCopy(context.TODO(), &s3.UploadPartCopyInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(objectName),
		PartNumber: int32(partNumberInt),
		UploadId:   aws.String(uploadId),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		log.Println("[Error]:[UploadPartCopy]", err, '\n')
		log.Println("[Error]:[UploadPartCopy] ObjectKey", objectName)
		// HandleError(err)
		return nil, nil, err
	}
	resXML, err := xml.MarshalIndent(resp.CopyPartResult, " ", " ")
	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()

	respGetRawResponse := middleware.GetRawResponse(resp.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	resMp["ETag"] = httpResponse.Header.Get("Etag")
	return resMp, res, err
}
