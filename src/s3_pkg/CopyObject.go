package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"log"
	"net/http"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) CopyObject(bucketName, objectName,
	copySource string,header http.Header) (map[string]string, []byte, error) {
	err := pkg.GetS3Client()
	if err != nil {
		Logger.Info("CopyObject", zap.Any("error", err))
		return nil, nil, err
	}
	copyObjectInput := &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(objectName),
		CopySource: aws.String(copySource),
	}
	if header.Get("Content-Type") != "" {
		copyObjectInput.ContentType = aws.String(header.Get("Content-Type"))
	}
	if header.Get("X-Oss-Server-Side-Encryption") != ""{
		copyObjectInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Oss-Server-Side-Encryption"))
	}
	if header.Get("X-Amz-Server-Side-Encryption") != ""{
		copyObjectInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Amz-Server-Side-Encryption"))
	}
	if header.Get("X-Oss-Object-Acl") != "" {
		copyObjectInput.ACL = types.ObjectCannedACL(header.Get("X-Oss-Object-Acl"))
	}
	if header.Get("X-Amz-Object-Acl") != "" {
		copyObjectInput.ACL = types.ObjectCannedACL(header.Get("X-Amz-Object-Acl"))
	}
	resp, err := pkg.S3Client.CopyObject(context.TODO(), copyObjectInput)
	if err != nil {
		log.Println("[Error]:[PutObject]", err, '\n')
		log.Println("[Error]:[PutObject] ObjectKey", objectName)
		// HandleError(err)
		return nil, nil, err
	}

	resXML, err := xml.MarshalIndent(resp.CopyObjectResult, " ", " ")
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
