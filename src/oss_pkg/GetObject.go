// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Zijun Hu, Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"io"
	"log"
)

func (pkg *OssPkg) GetObject(
	accessKeyId string, accessKeySecret string,
	endpoint string, bucketName string, objectName string,
	paramsMap map[string]string) ([]byte, error) {

	err := pkg.GetOssClient()
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

	goReadCloser, err := bucket.GetObject(objectName)
	if err != nil {
		log.Println("[Error]:[bucket.GetObject]", err, '\n')
		log.Println("[Error]:[bucket.GetObject] ObjectKey", objectName)
		// HandleError(err)
		return nil, err
	}
	log.Println("res result is", goReadCloser)
	bytes, err := io.ReadAll(goReadCloser)
	if err != nil {
		log.Println("[Error]:[bucket.GetObject] read error", err, '\n')
		// HandleError(err)
		return nil, err
	}
	// log.Println("GetObject Done!")

	return bytes, nil
}
