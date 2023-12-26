// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Zijun Hu, Chao Qin
// ///////////////////////////////////////
package s3_pkg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) GetObject(
	bucketName string, objectName string,
	urlQuery url.Values, urlHeader http.Header,
) (respHeader map[string]string, respBody io.ReadCloser, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.GetObject", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}
	if GetQueryString(urlQuery, "versionId") != "" {
		versionId := GetQueryString(urlQuery, "versionId")
		s3GetObjectInput.VersionId = &versionId
	}
	if GetQueryString(urlQuery, "response-cache-control") != "" {
		responseCacheControl := GetQueryString(urlQuery, "response-cache-control")
		s3GetObjectInput.ResponseCacheControl = &responseCacheControl
	}
	if GetQueryString(urlQuery, "response-content-disposition") != "" {
		responseContentDisposition := GetQueryString(urlQuery, "response-content-disposition")
		s3GetObjectInput.ResponseContentDisposition = &responseContentDisposition
	}
	if GetQueryString(urlQuery, "response-content-encoding") != "" {
		responseContentEncoding := GetQueryString(urlQuery, "response-content-encoding")
		s3GetObjectInput.ResponseContentEncoding = &responseContentEncoding
	}
	if GetQueryString(urlQuery, "response-content-language") != "" {
		responseContentLanguage := GetQueryString(urlQuery, "response-content-language")
		s3GetObjectInput.ResponseContentLanguage = &responseContentLanguage
	}
	if GetQueryString(urlQuery, "response-expires") != "" {
		strResponseExpires := GetQueryString(urlQuery, "response-expires")
		timeLayout := "2006-01-02 15:04:05"
		loc, _ := time.LoadLocation("local")
		responseExpires, _ := time.ParseInLocation(timeLayout, strResponseExpires, loc)
		s3GetObjectInput.ResponseExpires = &responseExpires
	}
	if urlHeader.Get("Range") != "" {
		s3Range := urlHeader.Get("Range")
		s3GetObjectInput.Range = &s3Range
	}

	if urlHeader.Get("If-Modified-Since") != "" {
		modTime := urlHeader.Get("If-Modified-Since")
		t, _ := time.Parse(time.RFC1123, modTime)
		t = t.UTC()
		s3GetObjectInput.IfModifiedSince = aws.Time(t)
	}

	if urlHeader.Get("If-Unmodified-Since") != "" {
		modTime := urlHeader.Get("If-Unmodified-Since")
		t, _ := time.Parse(time.RFC1123, modTime)
		t = t.UTC()
		s3GetObjectInput.IfUnmodifiedSince = aws.Time(t)
	}

	if urlHeader.Get("If-Match") != "" {
		tag := urlHeader.Get("If-Match")
		s3GetObjectInput.IfMatch = aws.String(tag)
	}

	if urlHeader.Get("If-None-Match") != "" {
		tag := urlHeader.Get("If-None-Match")
		s3GetObjectInput.IfNoneMatch = aws.String(tag)
	}

	s3GoRes, err := pkg.S3Client.GetObject(context.TODO(), s3GetObjectInput)
	if err != nil {
		Logger.Error("S3Pkg.GetObject ",
			zap.Any("bucketName", bucketName),
			zap.Any("objectName", objectName),
			zap.Any("error", err))
		// HandleError(err)
		return nil, nil, err
	}

	// Header
	respGetRawResponse := mid.GetRawResponse(s3GoRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	resMp["Content-Length"] = fmt.Sprintf("%d", s3GoRes.ContentLength)
	// Body
	/*defer s3GoRes.Body.Close()
	ioReader, ioErr := io.ReadAll(s3GoRes.Body)
	if ioErr != nil {
		log.Println("[Error]:[S3Client.GetObject] read error", ioErr, '\n')
		// HandleError(err)
		return nil, nil, ioErr
	}
	respBody = ioReader

	log.Println("[xxxx]:[S3Client.GetObject] respHeader:", respHeader)

	return respHeader, respBody, nil*/
	return resMp, s3GoRes.Body, nil
}
