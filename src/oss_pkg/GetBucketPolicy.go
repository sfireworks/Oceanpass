package oss_pkg

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
	"net/http"
)

func (pkg *OssPkg) GetBucketPolicy(bucketName string) (res []byte, resmap map[string]string, err error) {
	err = pkg.GetOssClient()
	if err != nil {
		Logger.Error("<OssPkg>.GetBucketPolicy(): [New] GetOssClient(): ",
			zap.Any("error", err))
		return nil, nil, err
	}
	var respHeader http.Header
	gbpRes, err := pkg.Client.GetBucketPolicy(bucketName, oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("<OssPkg>.GetBucketPolicy()",
			zap.String("bucketName", bucketName),
			zap.Any("error", err))
		return nil, nil, err
	}
	Logger.Info("<OssPkg>.GetBucketPolicy(): ok.", zap.Any("result", gbpRes))
	//log.Println("GetBucketPolicy result:", gbpRes)

	headResp := []byte(gbpRes)
	resMp := make(map[string]string)
	for resKey, resVal := range respHeader {
		resTempVal := ""
		for _, val := range resVal {
			resTempVal += val
		}
		resMp[resKey] = resTempVal
	}
	resMp["Etag"] = respHeader.Get("ETag")
	resMp["Content-Length"] = fmt.Sprintf("%d", len(headResp))
	return headResp, resMp, nil
}
