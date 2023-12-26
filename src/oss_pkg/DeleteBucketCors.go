package oss_pkg

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
	"net/http"
)

func (pkg *OssPkg) DeleteBucketCors(
	bucketName string) (map[string]string, error) {
	err := pkg.GetOssClient()
	if err != nil {
		Logger.Error("OssPkg.DeleteBucketCors(): [New] GetOssClient(): ",
			zap.Any("error", err))
		return nil, err
	}
	// Delete cors
	var respHeader http.Header
	deleteBucketCorsErr := pkg.Client.DeleteBucketCORS(bucketName, oss.GetResponseHeader(&respHeader))
	if deleteBucketCorsErr != nil {
		Logger.Error("OssPkg.DeleteBucketCors()",
			zap.String("bucketName", bucketName),
			zap.Any("error", deleteBucketCorsErr))
		return nil, deleteBucketCorsErr
	}
	// deleteBucketCorsErr is nil
	resMp := make(map[string]string)
	for resKey, resVal := range respHeader {
		resTempVal := ""
		for _, val := range resVal {
			resTempVal += val
		}
		resMp[resKey] = resTempVal
	}
	resMp["Etag"] = respHeader.Get("ETag")
	Logger.Info("OssPkg.DeleteBucketCors(): call oss.DeleteBucketCors() ok.")
	return resMp, nil
}
