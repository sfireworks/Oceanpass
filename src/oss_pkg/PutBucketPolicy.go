package oss_pkg

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

func (pkg *OssPkg) PutBucketPolicy(
	bucketName string, r *http.Request) error {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	policyInfo := string(bodyBytes)
	err := pkg.GetOssClient()

	if err != nil {
		Logger.Error("<OssPkg>.PutBucketPolicy(): [New] GetOssClient(): ", zap.Any("error", err))
		HandleError(err)
		return err
	}
	// Set policy
	Logger.Info("<OssPkg>.PutBucketPolicy() input: ",
		zap.String("accessKeyId", pkg.AccessKeyId),
		zap.String("accessKeySecret", pkg.AccessKeySecret),
		zap.String("endpoint", pkg.Endpoint),
		zap.String("bucketName", bucketName),
		zap.Any("policyInfo", policyInfo))
	var respHeader http.Header
	err = pkg.Client.SetBucketPolicy(bucketName, policyInfo, oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("<OssPkg>.PutBucketPolicy()",
			zap.String("accessKeyId", pkg.AccessKeyId),
			zap.String("accessKeySecret", pkg.AccessKeySecret),
			zap.String("endpoint", pkg.Endpoint),
			zap.String("bucketName", bucketName),
			zap.Any("error", err))
		return err
	}
	Logger.Info("<OssPkg>.PutBucketPolicy(): call oss.SetBucketPolicy() ok.")
	resMp := make(map[string]string)
	for resKey, resVal := range respHeader {
		resTempVal := ""
		for _, val := range resVal {
			resTempVal += val
		}
		resMp[resKey] = resTempVal
	}
	resMp["Etag"] = respHeader.Get("ETag")
	return nil
}
