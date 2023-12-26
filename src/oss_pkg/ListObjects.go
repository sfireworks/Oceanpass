// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func getObjectsFormResponse(lores oss.ListObjectsResult) string {
	var output string
	for _, object := range lores.Objects {
		output += object.Key + "  "
	}
	return output
}

type ListObjectsReq struct {
	Delimiter string `json:"delimiter"`
	Marker    string `json:"marker"`
	Maxkeys   int64  `json:"max-keys"`
	Prefix    string `json:"prefix"`
}

// check the paramters' validity
func CheckListObjectsParams(params ListObjectsReq) (err error) {
	//log.Println(params)
	return
}

func GetHttpQueryParams(vars url.Values) (map[string]string, bool) {
	mp := make(map[string]string)
	mp["delimiter"] = GetQueryString(vars, "delimiter")
	mp["marker"] = GetQueryString(vars, "marker")
	mp["max-keys"] = "1000"
	if GetQueryString(vars, "max-keys") != "" {
		max_keys := GetQueryString(vars, "max-keys")
		intMaxkeys, _ := strconv.Atoi(max_keys)
		if intMaxkeys >= 2 || intMaxkeys <= 1000 {
			mp["max-keys"] = max_keys
		}
	}
	mp["prefix"] = GetQueryString(vars, "prefix")

	flag := false
	for k, _ := range mp {
		if mp[k] != "" {
			flag = true
		}
	}
	return mp, flag
}

func (pkg *OssPkg) ListObjects(
	bucketName string,
	paramsMap map[string]string) ([]byte, map[string]string, error) {

	log.Println(
		"[OssPkg] Call ListObjects.",
		"accessKeyId :", pkg.AccessKeyId,
		" accessKeySecret:", pkg.AccessKeySecret,
		" endpoint:", pkg.Endpoint,
		"\n bucketName:", bucketName,
		" params:", paramsMap,
		"\n ")

	err := pkg.GetOssClient()
	if err != nil {
		log.Println("[Error]:[New]", err)
		HandleError(err)
		return nil, nil, err
	}
	var respHeader http.Header
	bucket, err := pkg.Client.Bucket(bucketName)
	if err != nil {
		log.Println("[Error]:[client.Bucket]", err)
		HandleError(err)
		return nil, nil, err
	}

	delimiter := paramsMap["delimiter"]
	ossDelimiter := oss.Delimiter(delimiter)

	marker := paramsMap["marker"]
	ossMarker := oss.Marker(marker)

	max_keys := paramsMap["max-keys"]
	maxkeys, _ := strconv.Atoi(max_keys)
	ossMaxKeys := oss.MaxKeys(maxkeys)

	prefix := paramsMap["prefix"]
	ossPrefix := oss.Prefix(prefix)

	loRes, err := bucket.ListObjects(
		ossDelimiter, ossMarker, ossMaxKeys, ossPrefix, oss.GetResponseHeader(&respHeader))
	if err != nil {
		log.Println("[Error]:[bucket.ListObjects]", err)
		HandleError(err)
		return nil, nil, err
	}

	resXML, err := xml.MarshalIndent(loRes, " ", " ")
	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
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
