// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

type GetBucketInfoReq struct {
}

func (pkg *OssPkg) GetBucketInfo(
	bucketName string,
	urlQuery url.Values) ([]byte, map[string]string, error) {

	err := pkg.GetOssClient()
	if err != nil {
		Logger.Error("<oss>.GetBucketInfo() [New]:", zap.Any("error", err))
		HandleError(err)
		return nil, nil, err
	}
	var respHeader http.Header
	// TODO: the second argument option is not filled yet.
	gBRes, err := pkg.Client.GetBucketInfo(bucketName, nil, oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("<oss>.GetBucketInfo() ", zap.Any("error", err))
		HandleError(err)
		return nil, nil, err
	}

	resXML, err := xml.MarshalIndent(gBRes, " ", " ")
	if err != nil {
		Logger.Error(
			"<oss>.GetBucketInfo(): marshal xml err ",
			zap.Any("error", err))
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
	return res, resMp, nil
}
