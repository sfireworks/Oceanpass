package oss_pkg

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
	"net/http"
	"oceanpass/src/common"
)

func (pkg *OssPkg) GetBucketACL(bucketName string) (
	bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetOssClient()
	if err != nil {
		Logger.Error("OssPkg.GetBucketACL(): [New] GetOssClient(): ",
			zap.Any("error", err))
		return nil, nil, err
	}
	var respHeader http.Header
	getBucketACLRes, getBucketACLErr := pkg.Client.GetBucketACL(bucketName, oss.GetResponseHeader(&respHeader))
	if getBucketACLErr != nil {
		Logger.Error("OssPkg.GetBucketACL()",
			zap.String("bucketName", bucketName),
			zap.Any("error", getBucketACLErr))
		return nil, nil, getBucketACLErr
	}

	res, err := common.MarshalIndent(getBucketACLRes)
	if err != nil {
		Logger.Error("OssPkg.GetBucketACL(): marshal xml err ",
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
	Logger.Info("OssPkg.GetBucketACL(): ok.",
		zap.Any("result", string(res)))
	return res, resMp, nil
}
