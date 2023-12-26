package s3_pkg

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) UploadPart(bucketName string, objectName string,
	partNumber, uploadId string, contentLength int64, header http.Header,
	data io.Reader) (map[string]string, error) {

	err := pkg.GetS3Client()
	if err != nil {
		Logger.Info("UploadPart", zap.Any("error", err))
		return nil, err
	}
	partNumberInt, _ := strconv.ParseInt(partNumber, 10, 32)
	uploadPartInput := &s3.UploadPartInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectName),
		PartNumber:    int32(partNumberInt),
		UploadId:      aws.String(uploadId),
		Body:          data,
		ContentLength: contentLength,
	}
	if header.Get("Content-Md5") != ""{
		uploadPartInput.ContentMD5 = aws.String(header.Get("Content-Md5"))
	}
	resp, err := pkg.S3Client.UploadPart(context.TODO(), uploadPartInput, s3.WithAPIOptions(
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	if err != nil {
		log.Println("[Error]:[UploadPart]", err, '\n')
		log.Println("[Error]:[UploadPart] ObjectKey", objectName)
		// HandleError(err)
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(resp.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	resMp["ETag"] = *resp.ETag
	log.Println("UploadPart Header Map is", resMp)

	return resMp, err

}
