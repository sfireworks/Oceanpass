// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"log"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type ListObjectsV2Req struct {
	Delimiter string `json:"delimiter"`
	Marker    string `json:"marker"`
	Maxkeys   int64  `json:"max-keys"`
	Prefix    string `json:"prefix"`
}

func (pkg *OssPkg) ListObjectsV2(
	bucketName string,
	paramsMap map[string]string) ([]byte, error) {

	log.Println(
		"[OssPkg] Call ListObjectsV2.",
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
		return nil, err
	}

	bucket, err := pkg.Client.Bucket(bucketName)
	if err != nil {
		log.Println("[Error]:[client.Bucket]", err)
		HandleError(err)
		return nil, err
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

	list_type := paramsMap["list-type"]
	listtype, _ := strconv.Atoi(list_type)
	ossListType := oss.ListType(listtype)

	continuation_token := paramsMap["continuation-token"]
	ossContinuationToken := oss.ContinuationToken(continuation_token)

	fetch_owner := paramsMap["fetch-owner"]
	bool_fetch_owner, _ := strconv.ParseBool(fetch_owner)
	ossFetchOwner := oss.FetchOwner(bool_fetch_owner)

	start_after := paramsMap["start-after"]
	ossStartAfter := oss.StartAfter(start_after)

	lov2Res, err := bucket.ListObjectsV2(
		ossDelimiter, ossMarker, ossMaxKeys, ossPrefix,
		ossListType, ossContinuationToken,
		ossFetchOwner, ossStartAfter)
	if err != nil {
		log.Println("[Error]:[bucket.ListObjectsV2]", err)
		// HandleError(err)
		return nil, err
	}

	resXML, err := xml.MarshalIndent(lov2Res, " ", " ")
	log.Println("lov2Res: ", lov2Res)

	if err != nil {
		log.Printf("marshal xml err :%v\n", err)
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()

	return res, nil
}
