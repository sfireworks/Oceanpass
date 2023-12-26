package oss_pkg

import (
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

func (pkg *OssPkg) DeleteBucketPolicy(
	bucketName string) (http.Header, error) {
	err := pkg.GetOssClient()
	if err != nil {
		Logger.Error("<OssPkg>.DeleteBucketPolicy(): [New] GetOssClient(): ",
			zap.Any("error", err))
		return nil, err
	}
	// Delete policy
	var respHeader http.Header
	err = pkg.Client.DeleteBucketPolicy(bucketName, oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("<OssPkg>.DeleteBucketPolicy()",
			zap.String("accessKeyId", pkg.AccessKeyId),
			zap.String("accessKeySecret", pkg.AccessKeySecret),
			zap.String("endpoint", pkg.Endpoint),
			zap.String("bucketName", bucketName),
			zap.Any("error", err))
		return nil, err
	}
	Logger.Info("<OssPkg>.DeleteBucketPolicy(): call oss.DeleteBucketPolicy() ok.",
		zap.Any("respHeader", respHeader))
	return respHeader, nil
}
