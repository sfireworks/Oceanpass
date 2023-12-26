// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"oceanpass/src/common"
	"sort"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

func (pkg *OssPkg) CompleteMultipartUpload(
	bucketName string, objectName string,
	urlQuery url.Values, partNumEtag map[int]string,
	headerXOssCallback string, headerXOssCallbackVar string,
) (res []byte, err error) {
	// Init Oss Bucket with bucketName
	err = pkg.GetOssBucket(bucketName)
	if err != nil {
		Logger.Error("OssPkg.CompleteMultipartUpload():GetOssBucket error",
			zap.String("bucketName", bucketName), zap.Any("error", err))
		return nil, err
	}

	// params of CompleteMultipartUpload
	imur := oss.InitiateMultipartUploadResult{
		Bucket:   bucketName,
		Key:      objectName,
		UploadID: common.GetQueryString(urlQuery, "uploadId"),
	}
	SortedPartkeys := make([]int, 0, len(partNumEtag))
	for k, _ := range partNumEtag {
		SortedPartkeys = append(SortedPartkeys, k)
	}
	sort.Ints(SortedPartkeys)
	parts := []oss.UploadPart{}
	for _, partNum := range SortedPartkeys {
		strPartEtg := partNumEtag[partNum]
		parts = append(parts,
			oss.UploadPart{
				ETag:       strPartEtg,
				PartNumber: partNum,
			})
	}

	dataXOssCallback, errXOssCallback :=
		base64.StdEncoding.DecodeString(headerXOssCallback)
	dataXOssCallbackVar, errXOssCallbackVar :=
		base64.StdEncoding.DecodeString(headerXOssCallbackVar)
	if errXOssCallback != nil || errXOssCallbackVar != nil {
		Logger.Error("OssPkg.CompleteMultipartUpload(): ",
			zap.Any("errXOssCallback", errXOssCallback),
			zap.Any("errXOssCallbackVar", errXOssCallbackVar))
		err := fmt.Errorf("OssPkg.CompleteMultipartUpload(): "+
			"errXOssCallback(%+v), errXOssCallbackVar(%+v),",
			errXOssCallback, errXOssCallbackVar)
		return nil, err
	}
	strXOssCallback := string(dataXOssCallback)
	strXOssCallbackVar := string(dataXOssCallbackVar)
	Logger.Info("OssPkg.CompleteMultipartUpload():",
		zap.Any("http-XOssCallback", strXOssCallback),
		zap.Any("http-XOssCallbackVar", strXOssCallbackVar))

	// transfer string to base64(json)
	jsonXOssCallback := parseCallback(strXOssCallback)
	callbackBuffer := bytes.NewBuffer([]byte{})
	callbackEncoder := json.NewEncoder(callbackBuffer)
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(jsonXOssCallback)
	if err != nil {
		Logger.Error("OssPkg.CompleteMultipartUpload():callbackEncoder.Encode error",
			zap.String("endpoint", pkg.Endpoint), zap.String("bucketName", bucketName),
			zap.String("callbackVal", strXOssCallback),
			zap.Any("error", err))
		return nil, err
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())

	Logger.Info("OssPkg.CompleteMultipartUpload(): ",
		zap.String("bucketName", bucketName), zap.String("objectName", objectName),
		zap.Any("imur", imur), zap.Any("parts", parts),
		zap.Any("callbackBase64", callbackBase64))
	// call osssdk.CompleteMultipartUpload()
	var respHeader http.Header
	respCompleteMultipartUpload, err := pkg.Bucket.CompleteMultipartUpload(imur, parts,
		// TODO: use oss.Callback,  will return err(EOF)
		oss.Callback(callbackBase64),
		// TODO: oss.GetResponseHeader -> CallbackResult
		oss.GetResponseHeader(&respHeader))
	if err != nil {
		// "EOF" is a special treatment, which is associated with
		// the "EOF" error handling in oss_pkg.HandleErrorReturn().
		// After it is fixed, modify the conditions here.
		if err.Error() != "EOF" {
			Logger.Error("OssPkg.CompleteMultipartUpload() error",
				zap.String("bucketName", bucketName), zap.String("objectName", objectName),
				zap.Any("respCompleteMultipartUpload", respCompleteMultipartUpload),
				zap.Any("error", err))
			return nil, err
		}
	}

	// fake "{\"Status\":\"OK\"}" due to the incomplete function of Alibaba Cloud SDK
	// TODO: remove fake return.
	strResponseBody := ""
	if err.Error() == "EOF" {
		strResponseBody = "{\"Status\":\"OK\"}"
		var buffer bytes.Buffer
		buffer.Write([]byte(strResponseBody))
		res = buffer.Bytes()
		Logger.Info("OssPkg.CompleteMultipartUpload(): ok. ", zap.Any("respHeader", respHeader),
			zap.Any("respCompleteMultipartUpload", respCompleteMultipartUpload))
		return res, nil
	}

	// TODO: after removing fake return, go below.
	resXML, err := xml.MarshalIndent(respCompleteMultipartUpload, " ", " ")
	if err != nil {
		Logger.Error("OssPkg.CompleteMultipartUpload(): marshal xml err ",
			zap.Any("error", err))
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res = buffer.Bytes()

	Logger.Info("OssPkg.CompleteMultipartUpload(): ok. ", zap.Any("respHeader", respHeader),
		zap.Any("respCompleteMultipartUpload", respCompleteMultipartUpload))

	return res, nil
}
