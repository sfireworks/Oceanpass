// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

type ListBucketsReq struct {
}

func (pkg *OssPkg) ListBuckets(
	bucketName string,
	paramsMap map[string]string) ([]byte, map[string]string, error) {

	err := pkg.GetOssClient()

	if err != nil {
		Logger.Error("<oss>.ListBuckets() [New]:", zap.Any("error", err))
		HandleError(err)
		return nil, nil, err
	}

	max_keys := paramsMap["max-keys"]
	maxkeys, _ := strconv.Atoi(max_keys)
	ossMaxKeys := oss.MaxKeys(maxkeys)

	Logger.Info("<oss>.ListBuckets()",
		zap.String("endpoint", pkg.AccessKeyId),
		zap.Any("ossMaxKeys", ossMaxKeys))
	var respHeader http.Header
	lbRes, err := pkg.Client.ListBuckets(ossMaxKeys, oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("<oss>.ListBuckets() ", zap.Any("error", err))
		HandleError(err)
		return nil, nil, err
	}

	resXML, err := xml.MarshalIndent(lbRes, " ", " ")
	if err != nil {
		Logger.Error("<oss>.ListBuckets(): marshal xml err ", zap.Any("error", err))
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	res := buffer.Bytes()
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
	// Logger.Info("<oss>.ListBuckets()  resXML string():", zap.Any("res", string(res))) // has '\n'

	return res, resMp, nil
}
