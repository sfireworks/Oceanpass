// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"log"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (pkg *OssPkg) GetObjectMeta(
	bucketName string, objectName string, versionId string,
) (headResp http.Header, err error) {
	err = pkg.GetOssClient()
	if err != nil {
		log.Println("[Error]:[New]", err)
		HandleError(err)
		return nil, err
	}
	bucket, err := pkg.Client.Bucket(bucketName)
	if err != nil {
		log.Println("[Error]:[client.Bucket]", err)
		HandleError(err)
		return nil, err
	}
	var gomRes http.Header
	if versionId != "" {
		gomRes, err = bucket.GetObjectMeta(objectName, oss.VersionId(versionId))
	} else {
		gomRes, err = bucket.GetObjectMeta(objectName)
	}
	if err != nil {
		log.Println("[Error]:[bucket.GetObjectMeta] ObjectKey", objectName, "\n err:", err, "\n ")
		// HandleError(err)
		return nil, err
	}
	log.Println("GetObjectMeta result:", gomRes)

	headResp = gomRes

	return headResp, nil
}
