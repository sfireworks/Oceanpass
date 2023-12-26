package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type DeleteResult struct {
	Deleted []types.DeletedObject
	Errors  []types.Error
	//noSmithyDocumentSerde
}

type DeleteObjectsInput struct {
	DelObjs []types.ObjectIdentifier `xml:"Object"`
	Quiet   bool                     `xml:"Quiet"`
}

func (s3pkg *S3Pkg) DeleteObjects(
	bucketName string, bodyBytes []byte,
) (res []byte, resMp map[string]string, err error) {
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.DeleteObjects ", zap.Any("error", err))
		return nil, resMp, err
	}

	DelObjsInput := DeleteObjectsInput{}
	err = xml.Unmarshal(bodyBytes, &DelObjsInput)

	if err != nil {
		println("Umarshal ERROR: ", err)
	}

	s3DelObjsInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: DelObjsInput.DelObjs,
			Quiet:   DelObjsInput.Quiet,
		},
	}

	s3DelObjsRes, err := s3pkg.S3Client.DeleteObjects(context.TODO(), s3DelObjsInput)
	if err != nil {
		Logger.Error("S3Pkg.DeleteObjects ",
			zap.Any("s3DelObjsInput", s3DelObjsInput),
			zap.Any("error", err))
		HandleError(err)
		return nil, resMp, err
	}

	respGetRawResponse := middleware.GetRawResponse(s3DelObjsRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp = make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}

	delObjsRes := DeleteResult{
		Deleted: s3DelObjsRes.Deleted,
		Errors:  s3DelObjsRes.Errors,
	}

	resXML, err := xml.MarshalIndent(delObjsRes, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.DeleteObjects(): marshal xml err ",
			zap.Any("error", err))
		return nil, resMp, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res = buffer.Bytes()

	/*
		The Content-Length in the common response header represents the exact
		length of the body that are filled into the response writer.
		The go http package can automatically calculate the body length under 2048 bytes,
		but when the body len is more than 2048 bytes, the length needs to be set manually in the code.
		That's why we can not get the correct response when deleting more than 21 objs, because the response
		length is more than 2048 bytes and we did not set the correct content length.
		Refs:https://stackoverflow.com/questions/75091383/why-does-golang-http-responsewriter-auto-add-content-length-if-its-no-more-than
		https://pkg.go.dev/net/http#ResponseWriter

	*/
	finLen := buffer.Len()
	resMp["Content-Length"] = strconv.Itoa(finLen)

	return res, resMp, nil
}
