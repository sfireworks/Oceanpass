// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Wenrui Yan
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

func (pkg *OssPkg) CopyObject(
	bucketName string, objectName string,
	r *http.Request) (resMp map[string]string, resBytes []byte, err error) {

	err = pkg.GetOssClient()
	if err != nil {
		Logger.Error("OssPkg.CopyObject()",
			zap.Any("GetClient", err))
		return nil, nil, err
	}

	bucket, err := pkg.Client.Bucket(bucketName)
	if err != nil {
		Logger.Error("OssPkg.CopyObject()",
			zap.Any("GetBucket", err))
		return nil, nil, err
	}

	var srcBucket, srcObject string
	srcBucket, srcObject = GetCopySourceAttr(r)

	options := []oss.Option{}
	ForbidOverWrite := r.Header.Get("X-Oss-Forbid-Overwrite")
	if ForbidOverWrite == "true" {
		options = append(options, oss.ForbidOverWrite(true))
	}
	if r.Header.Get("Content-Type") != "" {
		options = append(options, oss.ContentType(r.Header.Get("Content-Type")))
	}
	if r.Header.Get("X-Oss-Server-Side-Encryption") != "" {
		options = append(options, oss.ServerSideEncryption(r.Header.Get("X-Oss-Server-Side-Encryption")))
	}
	if r.Header.Get("X-Amz-Server-Side-Encryption") != "" {
		options = append(options, oss.ServerSideEncryption(r.Header.Get("X-Amz-Server-Side-Encryption")))
	}
	if r.Header.Get("Content-Md5") != "" {
		options = append(options, oss.ContentMD5(r.Header.Get("Content-Md5")))
	}
	if r.Header.Get("X-Oss-Object-Acl") != "" {
		options = append(options, oss.ACReqHeaders(r.Header.Get("X-Oss-Object-Acl")))
	}
	if r.Header.Get("X-Amz-Object-Acl") != "" {
		options = append(options, oss.ACReqHeaders(r.Header.Get("X-Amz-Object-Acl")))
	}
	Logger.Info("OssPkg.CopyObject()", zap.Any("srcBucket", srcBucket), zap.Any("srcObject", srcObject),
		zap.Any("bucketName", bucketName), zap.Any("objectName", objectName))
	var res oss.CopyObjectResult
	if bucketName == srcBucket {
		res, err = bucket.CopyObject(srcObject, objectName, options...)
	} else {
		if srcBucket == "" {
			Logger.Error("OssPkg.CopyObject():params error(miss srcBucket)")
			ocnErr := Error{
				HttpsResponseErrorStatusCode: "400",
				Code:                         "InvalidArgument",
				Message: "Copy Source must mention" +
					" the source bucket and key: /sourcebucket/sourcekey.",
				RequestId: "No RequestId due to parameter error (key[X-xxx-Copy-Source]).",
			}
			err = fmt.Errorf("oss: service returned error:"+
				" StatusCode=%+s,"+
				" ErrorCode=%+s,"+
				" ErrorMessage=%+s,"+
				" RequestId=%+s,",
				ocnErr.HttpsResponseErrorStatusCode,
				ocnErr.Code,
				ocnErr.Message,
				ocnErr.RequestId)
			return nil, nil, err
		}
		res, err = bucket.CopyObjectFrom(srcBucket, srcObject, objectName, options...)
	}

	if err != nil {
		Logger.Error("CopyObject()", zap.Any("error", err))
		return nil, nil, err
	}

	resXML, err := xml.MarshalIndent(res, " ", " ")

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	resBytes = buffer.Bytes()

	resMp = make(map[string]string)
	resMp["ETag"] = res.ETag

	return resMp, resBytes, nil
}
