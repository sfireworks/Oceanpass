package oss_pkg

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"oceanpass/src/common"
)

func (pkg *OssPkg) GetObjectACL(
	bucketName, objectName string,
	urlQuery url.Values) (
	bodyBytes []byte, headerMap map[string]string, err error) {
	// Init Oss Bucket with bucketName
	err = pkg.GetOssBucket(bucketName)
	if err != nil {
		Logger.Error("OssPkg.GetObjectACL():GetOssBucket error",
			zap.String("bucketName", bucketName),
			zap.String("objectName", objectName),
			zap.Any("urlQuery", urlQuery),
			zap.Any("error", err))
		return nil, nil, err
	}

	var GetObjectACLRes oss.GetObjectACLResult
	var GetObjectACLErr error
	var respHeader http.Header
	if GetQueryString(urlQuery, "versionId") != "" {
		versionId := GetQueryString(urlQuery, "versionId")
		GetObjectACLRes, GetObjectACLErr = pkg.Bucket.GetObjectACL(objectName,
			oss.VersionId(versionId), oss.GetResponseHeader(&respHeader))
	} else {
		GetObjectACLRes, GetObjectACLErr = pkg.Bucket.GetObjectACL(objectName, oss.GetResponseHeader(&respHeader))
	}
	if GetObjectACLErr != nil {
		Logger.Error("OssPkg.GetObjectACL()",
			zap.String("bucketName", bucketName),
			zap.String("objectName", objectName),
			zap.Any("urlQuery", urlQuery),
			zap.Any("error", GetObjectACLErr))
		return nil, nil, GetObjectACLErr
	}
	Logger.Info("OssPkg.GetObjectACL(): ok.",
		zap.Any("result", GetObjectACLRes))

	res, err := common.MarshalIndent(GetObjectACLRes)
	if err != nil {
		Logger.Error("OssPkg.GetObjectACL(): marshal xml err ",
			zap.Any("error", err))
	}
	resMp := make(map[string]string)
	for resKey, resVal := range respHeader {
		resTempVal := ""
		for _, val := range resVal {
			resTempVal += val
		}
		resMp[resKey] = resTempVal
	}
	resMp["Etag"] = respHeader.Get("ETag")
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	Logger.Info("OssPkg.GetObjectACL(): ok.",
		zap.Any("result", res))
	Logger.Info("OssPkg.GetObjectACL(): ok.",
		zap.Any("result", string(res)))

	return res, resMp, nil
}
