// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package oss_pkg

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"oceanpass/src/zaplog"
	"regexp"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssPkg struct {
	Client                    *oss.Client
	Bucket                    *oss.Bucket
	BucketName                string
	Region                    string
	ServiceName               string
	AccessKeyId               string
	AccessKeySecret           string
	Endpoint                  string
	CfgStsDefaultDurationSecs string
	StsToken                  string
}

var Logger = zaplog.Logger

func (pkg *OssPkg) New() {

}

func HandleError(err error) {
	// fmt.Println("Error:", err)
	// os.Exit(-1)
}

func GetQueryString(vars url.Values, key string) string {
	if len(vars[key]) > 0 {
		return vars[key][0]
	}
	return ""
}

// TODO we need to reduce the object construction in prod
func (pkg *OssPkg) GetOssClient() (err error) {
	client, err := oss.New(
		pkg.Endpoint, pkg.AccessKeyId, pkg.AccessKeySecret, oss.SecurityToken(pkg.StsToken),
	)
	pkg.Client = client
	return err
}

func (pkg *OssPkg) GetOssBucket(bucketName string) (err error) {
	err = pkg.GetOssClient()
	if err != nil {
		return err
	}
	bucket, err := pkg.Client.Bucket(bucketName)
	if err != nil {
		return err
	}
	pkg.Bucket = bucket
	return err
}

// //////////////////////////////////////////////
// Callback for PutObject, CompleteMultipartUpload
type Callback struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackHost     string `json:"callbackHost"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
	// TODO: handle CallbackBody and CallbackVar
	// CallbackBody     CallbackBody `json:"callbackBody"`
}

type CallbackBody struct {
	Bucket          string
	Object          string
	Etag            string
	Size            string
	MimeType        string
	ImageInfoHeight string
	ImageInfoWidth  string
	ImageInfoFormat string
}

// TODO: handle CallbackBody and CallbackVar
// type CallbackBody struct {
// 	Bucket          string `json:"bucket"`
// 	Object          string `json:"object"`
// 	Etag            string `json:"etag"`
// 	Size            string `json:"size"`
// 	MimeType        string `json:"mimeType"`
// 	ImageInfoHeight string `json:"imageInfo.height"`
// 	ImageInfoWidth  string `json:"imageInfo.width"`
// 	ImageInfoFormat string `json:"imageInfo.format"`
// }

// TODO: add CallbackVar afer new go sdk of aliyun
// type CallbackVar struct {
// 	XXX1 string
// 	XXX2 string
// }

// transfer string to json::Callback
func parseCallback(input string) Callback {
	bodyBytes := []byte(input)
	var resCallback Callback
	err := json.Unmarshal([]byte(bodyBytes), &resCallback)
	if err != nil {
		Logger.Error("parseCallback(): json unmarshal failed")
	}
	return resCallback
}

// End: Callback for PutObject, CompleteMultipartUpload
////////////////////////////////////////////////

// Get copy source bucket and object name, agnostic of type of client
func GetCopySourceAttr(r *http.Request) (bktName string, objName string) {
	ossCpSrcVal := r.Header.Get("X-Oss-Copy-Source")
	amzCpSrcVal := r.Header.Get("X-Amz-Copy-Source")
	ossRegExp := "^/([^/]+)/(.+)$"
	amzRegExp := "^([^/]+)/(.+)$"

	if len(ossCpSrcVal) > 1 {
		re, err := regexp.Compile(ossRegExp)
		if err != nil {
			log.Fatalln("regexp apiErr failed!")
		}
		matchArr := re.FindStringSubmatch(ossCpSrcVal)
		if len(matchArr) < 3 {
			Logger.Fatal("regexp apiErr len mismatch!")
		}
		bktName = matchArr[1]
		objName, _ = url.QueryUnescape(matchArr[2])
		return bktName, objName
	}
	if len(amzCpSrcVal) > 1 {
		re, err := regexp.Compile(amzRegExp)
		if err != nil {
			log.Fatalln("regexp apiErr failed!")
		}
		matchArr := re.FindStringSubmatch(amzCpSrcVal)
		if len(matchArr) < 3 {
			Logger.Fatal("regexp apiErr len mismatch!")
		}
		bktName = matchArr[1]
		objName, _ = url.QueryUnescape(matchArr[2])
		return bktName, objName
	}
	return "", ""
}
