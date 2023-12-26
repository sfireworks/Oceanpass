// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"log"
	"net/http"
)

func (pkg *OssPkg) GetObjectDetailedMeta(
	bucketName string, objectName string,
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

	godmRes, err := bucket.GetObjectDetailedMeta(objectName)
	if err != nil {
		log.Println("[Error]:[bucket.GetObjectDetailedMeta] ObjectKey", objectName, "\n err:", err, "\n ")
		// HandleError(err)
		return nil, err
	}
	log.Println("GetObjectDetailedMeta result is", godmRes)

	headResp = godmRes

	return headResp, nil
}
