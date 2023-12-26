package s3_pkg

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"net/http"
	"strings"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	awshttp "github.com/aws/smithy-go/transport/http"
	//"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	//"log"
)

func (pkg *S3Pkg) PutObject(bucketName string, objectName string,
	header http.Header, data io.Reader) (map[string]string, error) {
	err := pkg.GetS3Client()
	if err != nil {
		Logger.Info("PutObject", zap.Any("error", err))
		return nil, err
	}
	metaMp := make(map[string]string)
	for key, val := range header {
		if strings.HasPrefix(key, "X-Amz-Meta-") || strings.HasPrefix(key, "X-Oss-Meta-") {
			metaKey := key[len("X-Oss-Meta-"):]
			metaKey = strings.ToLower(metaKey)
			metaMp[metaKey] = val[0]
		}
	}
	s3PutObjectInput := s3.PutObjectInput{
		Bucket:   aws.String(bucketName),
		Key:      aws.String(objectName),
		Body:     data,
		Metadata: metaMp,
	}
	if header.Get("Content-Type") != "" {
		s3PutObjectInput.ContentType = aws.String(header.Get("Content-Type"))
	}
	if header.Get("X-Oss-Server-Side-Encryption") != ""{
		s3PutObjectInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Oss-Server-Side-Encryption"))
	}
	if header.Get("X-Amz-Server-Side-Encryption") != ""{
		s3PutObjectInput.ServerSideEncryption = types.ServerSideEncryption(header.Get("X-Amz-Server-Side-Encryption"))
	}
	if header.Get("Content-Md5") != ""{
		s3PutObjectInput.ContentMD5 = aws.String(header.Get("Content-Md5"))
	}
	if header.Get("X-Oss-Object-Acl") != "" {
		s3PutObjectInput.ACL = types.ObjectCannedACL(header.Get("X-Oss-Object-Acl"))
	}
	if header.Get("X-Amz-Object-Acl") != "" {
		s3PutObjectInput.ACL = types.ObjectCannedACL(header.Get("X-Amz-Object-Acl"))
	}
	// uploader := manager.NewUploader(pkg.S3Client)

	// resp, err := uploader.Upload(context.TODO(), &s3PutObjectInput)
	// if err != nil {
	// 	log.Println("[Error]:[PutObject]", err, '\n')
	// 	log.Println("[Error]:[PutObject] ObjectKey", objectName)
	// 	// HandleError(err)
	// 	return nil, err
	// }
	resp, err := pkg.S3Client.PutObject(context.TODO(), &s3PutObjectInput)
	Logger.Info("<AwsPkg>.uploader.Upload()",
		zap.String("endpoint", pkg.Endpoint),
		zap.String("bucketName", bucketName),
		zap.String("objectName", objectName),
		zap.Any("resp.ETag", *resp.ETag),
	)
	respGetRawResponse := middleware.GetRawResponse(resp.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	resMp["ETag"] = *resp.ETag
	return resMp, err

}