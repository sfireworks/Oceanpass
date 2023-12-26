// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

func (pkg *OssPkg) PutObject(
	bucketName string, objectName string,
	data io.Reader,
	headerXOssCallback string, headerXOssCallbackVar string,
	// pBody []byte,	// TODO: handle CallbackBody and CallbackVar
) (map[string]string, []byte, error) {
	// Init Oss Bucket with bucketName
	err := pkg.GetOssBucket(bucketName)
	if err != nil {
		Logger.Error("OssPkg.PutObject():GetOssBucket error",
			zap.String("bucketName", bucketName), zap.Any("error", err))
		return nil, nil, err
	}
	bucket := pkg.Bucket

	// Log: before base64 for callbackVal
	Logger.Info("OssPkg.PutObject(): ",
		zap.String("bucketName", bucketName), zap.Any("data", data),
		zap.String("callbackValue", headerXOssCallback),
		zap.String("callbackVar", headerXOssCallbackVar))

	dataXOssCallback, errXOssCallback :=
		base64.StdEncoding.DecodeString(headerXOssCallback)
	dataXOssCallbackVar, errXOssCallbackVar :=
		base64.StdEncoding.DecodeString(headerXOssCallbackVar)
	if errXOssCallback != nil || errXOssCallbackVar != nil {
		Logger.Error("OssPkg.PutObject():DecodeString error ",
			zap.Any("errXOssCallback", errXOssCallback),
			zap.Any("errXOssCallbackVar", errXOssCallbackVar))
		err := fmt.Errorf("OssPkg.PutObject(): "+
			"errXOssCallback(%+v), errXOssCallbackVar(%+v),",
			errXOssCallback, errXOssCallbackVar)
		return nil, nil, err
	}
	strXOssCallback := string(dataXOssCallback)
	strXOssCallbackVar := string(dataXOssCallbackVar)
	Logger.Info("OssPkg.PutObject(): ",
		zap.Any("http-XOssCallback", strXOssCallback),
		zap.Any("http-XOssCallbackVar", strXOssCallbackVar))

	jsonXOssCallback := parseCallback(strXOssCallback)
	callbackBuffer := bytes.NewBuffer([]byte{})
	callbackEncoder := json.NewEncoder(callbackBuffer)
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(jsonXOssCallback)
	if err != nil {
		Logger.Error("OssPkg.PutObject():callbackEncoder.Encode error",
			zap.String("endpoint", pkg.Endpoint), zap.String("bucketName", bucketName),
			zap.Any("data", data), zap.String("callbackValue", headerXOssCallback),
			zap.Any("error", err))
		return nil, nil, err
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())
	// Log: after base64 for callbackVal
	Logger.Info("OssPkg.PutObject(): ",
		zap.String("bucketName", bucketName), zap.Any("data", data),
		zap.Any("callbackBase64", callbackBase64))

	var respHeader http.Header
	// TODO: CallbackResult
	err = bucket.PutObject(objectName, data,
		oss.Callback(callbackBase64), oss.GetResponseHeader(&respHeader))
	if err != nil {
		Logger.Error("OssPkg.PutObject()",
			zap.String("bucketName", bucketName),
			zap.String("objectName", objectName),
			zap.Any("callbackBase64", callbackBase64),
			zap.Any("respHeader", respHeader),
			zap.Any("error", err))
		return nil, nil, err
	}
	Logger.Info("OssPkg.PutObject(): ok. ", zap.Any("respHeader", respHeader))
	resMp := make(map[string]string)
	for resKey, resVal := range respHeader {
		resTempVal := ""
		for _, val := range resVal {
			resTempVal += val
		}
		resMp[resKey] = resTempVal
	}
	resMp["Etag"] = respHeader.Get("ETag")

	// TODO: fake "{\"Status\":\"OK\"}" due to the incomplete function of Alibaba Cloud SDK
	strResponseBody := "{\"Status\":\"OK\"}"
	var buffer bytes.Buffer
	buffer.Write([]byte(strResponseBody))
	resBytes := buffer.Bytes()

	return resMp, resBytes, nil
}

// TODO: add sth afer new go sdk of aliyun
// TODO: opCallbackResult := CallbackResult(&pBody)
