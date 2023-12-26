package encryption

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"hash"
	"io"
	"net/http"
	"net/url"
	"oceanpass/src/http_handler"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"sort"
	"strings"
	"time"
)

var (
	HTTPHeaderContentMD5  = "Content-MD5"
	HTTPHeaderContentType = "Content-Type"
	HTTPHeaderDate        = "Date"
)

// Copy from OSS SDK (oss/conn.go)
var signKeyList = []string{"acl", "uploads", "location", "cors",
	"logging", "website", "referer", "lifecycle",
	"delete", "append", "tagging", "objectMeta",
	"uploadId", "partNumber", "security-token",
	"position", "img", "style", "styleName",
	"replication", "replicationProgress",
	"replicationLocation", "cname", "bucketInfo",
	"comp", "qos", "live", "status", "vod",
	"startTime", "endTime", "symlink",
	"x-oss-process", "response-content-type", "x-oss-traffic-limit",
	"response-content-language", "response-expires",
	"response-cache-control", "response-content-disposition",
	"response-content-encoding", "udf", "udfName", "udfImage",
	"udfId", "udfImageDesc", "udfApplication", "comp",
	"udfApplicationLog", "restore", "callback", "callback-var", "qosInfo",
	"policy", "stat", "encryption", "versions", "versioning", "versionId", "requestPayment",
	"x-oss-request-payer", "sequential",
	"inventory", "inventoryId", "continuation-token", "asyncFetch",
	"worm", "wormId", "wormExtend", "withHashContext",
	"x-oss-enable-md5", "x-oss-enable-sha1", "x-oss-enable-sha256",
	"x-oss-hash-ctx", "x-oss-md5-ctx", "transferAcceleration",
	"regionList", "cloudboxes", "x-oss-ac-source-ip", "x-oss-ac-subnet-mask", "x-oss-ac-vpc-id", "x-oss-ac-forward-allow",
	"metaQuery",
}

// Copy from OSS SDK (oss/auth.go)
type headerSorter struct {
	Keys []string
	Vals []string
}

// Copy from OSS SDK (oss/auth.go/getSignedStr)
func CalculateOSSSignature(req *http.Request, keySecret string, version string, isPreSign bool, expires string) string {
	bucketName := http_handler.GetBucketName(req)
	objName := http_handler.GetObjectName(req, bucketName)
	subsource := getSubResource(req, version)
	canonicalizedResource := getResource(bucketName, objName, subsource, version)
	ossHeadersMap := make(map[string]string)
	additionalList, additionalMap := getAdditionalHeaderKeys(req)
	for k, v := range req.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-oss-") {
			ossHeadersMap[strings.ToLower(k)] = v[0]
		} else if version == "OSS2" {
			if _, ok := additionalMap[strings.ToLower(k)]; ok {
				ossHeadersMap[strings.ToLower(k)] = v[0]
			}
		}
	}
	hs := newHeaderSorter(ossHeadersMap)

	// Sort the ossHeadersMap by the ascending order
	hs.Sort()

	// Get the canonicalizedOSSHeaders
	canonicalizedOSSHeaders := ""
	for i := range hs.Keys {
		canonicalizedOSSHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}
	// Give other parameters values
	// when sign URL, date is expires
	date := req.Header.Get(HTTPHeaderDate)
	if isPreSign {
		date = expires
	}
	contentType := req.Header.Get(HTTPHeaderContentType)
	contentMd5 := req.Header.Get(HTTPHeaderContentMD5)
	// default is v1 signature
	signStr := req.Method + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedOSSHeaders + canonicalizedResource
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(keySecret))
	// v2 signature
	if version == "OSS2" {
		signStr = req.Method + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedOSSHeaders + strings.Join(additionalList, ";") + "\n" + canonicalizedResource
		h = hmac.New(func() hash.Hash { return sha256.New() }, []byte(keySecret))
	}
	//fmt.Println("----signStr---", signStr)
	io.WriteString(h, signStr)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	//fmt.Println("----signedStr---", signedStr)
	return signedStr
}

func PreSignCheck(secret, Expires string, r *http.Request) string {
	uri := r.URL.Path
	SubResource := getSubResource(r, "OSS")
	BucketName, ObjectName := "", ""
	BucketName = strings.Split(uri, "/")[1]
	if len(strings.Split(uri, "/")) > 2 {
		arr := strings.Split(uri, "/")[2:]
		ObjectName = strings.Join(arr, "/")
	}
	canonicalizedOSSHeaders := ""
	CanonicalizedResource := getResource(BucketName, ObjectName, SubResource, "OSS")

	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(secret))
	io.WriteString(h, r.Method+"\n"+r.Header.Get("Content-MD5")+"\n"+r.Header.Get("Content-Type")+"\n"+Expires+"\n"+canonicalizedOSSHeaders+CanonicalizedResource)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signedStr
}

// Copy from OSS SDK (oss/auth.go/getSignedStrV4)
func CalculateOSSV4Signature(req *http.Request, keySecret string, version string) string {
	// Find out the "x-oss-"'s address in header of the request
	bucketName := http_handler.GetBucketName(req)
	objName := http_handler.GetObjectName(req, bucketName)
	subsource := getSubResource(req, version)
	canonicalizedResource := getResourceV4(bucketName, objName, subsource)
	canonicalizedResource, err := url.QueryUnescape(canonicalizedResource)
	if err != nil {
		Logger.Error("GetSignedStr()", zap.Any("err", err))
		return ""
	}
	ossHeadersMap := make(map[string]string)
	for k, v := range req.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-oss-") {
			ossHeadersMap[strings.ToLower(k)] = strings.Trim(v[0], " ")
		}
	}

	// Required parameters
	signDate := ""
	dateFormat := ""
	date := req.Header.Get(HTTPHeaderDate)
	if date != "" {
		signDate = date
		dateFormat = http.TimeFormat
	}

	ossDate := req.Header.Get("X-Oss-Date")
	_, ok := ossHeadersMap[strings.ToLower("X-Oss-Date")]
	if ossDate != "" {
		signDate = ossDate
		dateFormat = "2006-01-02T15:04:05Z"
		if !ok {
			ossHeadersMap[strings.ToLower("X-Oss-Date")] = strings.Trim(ossDate, " ")
		}
	}

	contentType := req.Header.Get(HTTPHeaderContentType)
	_, ok = ossHeadersMap[strings.ToLower(HTTPHeaderContentType)]
	if contentType != "" && !ok {
		ossHeadersMap[strings.ToLower(HTTPHeaderContentType)] = strings.Trim(contentType, " ")
	}

	contentMd5 := req.Header.Get(HTTPHeaderContentMD5)
	_, ok = ossHeadersMap[strings.ToLower(HTTPHeaderContentMD5)]
	if contentMd5 != "" && !ok {
		ossHeadersMap[strings.ToLower(HTTPHeaderContentMD5)] = strings.Trim(contentMd5, " ")
	}
	hs := newHeaderSorter(ossHeadersMap)

	// Sort the ossHeadersMap by the ascending order
	hs.Sort()
	// Get the canonicalizedOSSHeaders
	canonicalizedOSSHeaders := ""
	for i := range hs.Keys {
		canonicalizedOSSHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	signStr := ""

	// v4 signature
	hashedPayload := req.Header.Get("X-Oss-Content-Sha256")

	// subResource
	resource := canonicalizedResource
	subResource := ""
	subPos := strings.LastIndex(canonicalizedResource, "?")
	if subPos != -1 {
		subResource = canonicalizedResource[subPos+1:]
		resource = canonicalizedResource[0:subPos]
	}

	// get canonical request
	canonicalReuqest := req.Method + "\n" + resource + "\n" + subResource + "\n" + canonicalizedOSSHeaders + "\n" + "" + "\n" + hashedPayload
	rh := sha256.New()
	io.WriteString(rh, canonicalReuqest)
	hashedRequest := hex.EncodeToString(rh.Sum(nil))

	// get day,eg 20210914
	t, _ := time.Parse(dateFormat, signDate)
	strDay := t.Format("20060102")

	str := http_handler.GetCredentialWithoutAccessKey(req)

	signStr = "OSS4-HMAC-SHA256" + "\n" + signDate + "\n" + str + "\n" + hashedRequest

	h1 := hmac.New(func() hash.Hash { return sha256.New() }, []byte("aliyun_v4"+keySecret))
	io.WriteString(h1, strDay)
	h1Key := h1.Sum(nil)
	signedStrV4Product := http_handler.GetServiceName(req)
	signedStrV4Region := http_handler.GetRegion(req)
	h2 := hmac.New(func() hash.Hash { return sha256.New() }, h1Key)
	io.WriteString(h2, signedStrV4Region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(func() hash.Hash { return sha256.New() }, h2Key)
	io.WriteString(h3, signedStrV4Product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(func() hash.Hash { return sha256.New() }, h3Key)
	io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	h := hmac.New(func() hash.Hash { return sha256.New() }, h4Key)
	io.WriteString(h, signStr)
	/*fmt.Println("---ossHeadersMap---",ossHeadersMap)
	fmt.Println("---canonicalReuqest---",canonicalReuqest)
	fmt.Println("---signStr---",signStr)
	fmt.Println("---res---",fmt.Sprintf("%x", h.Sum(nil)))*/
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Copy from OSS SDK (oss/conn.go)
func getResource(bucketName, objectName, subResource string, version string) string {
	if subResource != "" {
		subResource = "?" + subResource
	}
	if bucketName == "" {
		if version == "OSS2" {
			return url.QueryEscape("/") + subResource
		}
		return fmt.Sprintf("/%s%s", bucketName, subResource)
	}
	if version == "OSS2" {
		return url.QueryEscape("/"+bucketName+"/") + strings.Replace(url.QueryEscape(objectName), "+", "%20", -1) + subResource
	}
	return fmt.Sprintf("/%s/%s%s", bucketName, objectName, subResource)
}

// Copy from OSS SDK (oss/conn.go)
func getResourceV4(bucketName, objectName, subResource string) string {
	if subResource != "" {
		subResource = "?" + subResource
	}

	if bucketName == "" {
		return fmt.Sprintf("/%s", subResource)
	}

	if objectName != "" {
		objectName = url.QueryEscape(objectName)
		objectName = strings.Replace(objectName, "+", "%20", -1)
		objectName = strings.Replace(objectName, "%2F", "/", -1)
		return fmt.Sprintf("/%s/%s%s", bucketName, objectName, subResource)
	}
	return fmt.Sprintf("/%s/%s", bucketName, subResource)
}

// Copy from OSS SDK (oss/auth.go)
func newHeaderSorter(m map[string]string) *headerSorter {
	hs := &headerSorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

// Copy from OSS SDK (oss/auth.go)
func (hs *headerSorter) Sort() {
	sort.Sort(hs)
}

// Copy from OSS SDK (oss/auth.go)
func (hs *headerSorter) Len() int {
	return len(hs.Vals)
}

// Copy from OSS SDK (oss/auth.go)
func (hs *headerSorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

// Copy from OSS SDK (oss/auth.go)
func (hs *headerSorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}

// Copy from OSS SDK (oss/conn.go)
func getSubResource(req *http.Request, version string) string {
	// Sort
	params, _ := GetRawParams(req)
	fmt.Println("--params--",params)
	keys := make([]string, 0, len(params))
	signParams := make(map[string]string)
	for k := range params {
		if version == "OSS2" || version == "OSS4" {
			encodedKey := url.QueryEscape(k)
			keys = append(keys, encodedKey)
			if params[k] != nil && params[k] != "" {
				signParams[encodedKey] = strings.Replace(url.QueryEscape(params[k].(string)), "+", "%20", -1)
			}
		} else if isParamSign(k) {
			keys = append(keys, k)
			if params[k] != nil {
				signParams[k] = params[k].(string)
			}
		}
	}
	sort.Strings(keys)

	// Serialize
	var buf bytes.Buffer
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		if _, ok := signParams[k]; ok {
			if signParams[k] != "" {
				buf.WriteString("=" + signParams[k])
			}
		}
	}
	return buf.String()
}

// Copy from OSS SDK (oss/conn.go)
func isParamSign(paramKey string) bool {
	for _, k := range signKeyList {
		if paramKey == k {
			return true
		}
	}
	return false
}

// Copy from OSS SDK (oss/option.go)
func GetRawParams(req *http.Request) (map[string]interface{}, error) {
	// Option
	str := req.URL.Query()
	paramsm := map[string]interface{}{}
	// Serialize
	for k, v := range str {
		if k == "Signature" || k == "AccessKey" || k == "Expires" || k == "x-oss-signature"  {
			continue
		}
		if v[0] == "" {
			paramsm[k] = nil
		} else {
			paramsm[k] = v[0]
		}
	}
	return paramsm, nil
}

func getAdditionalHeaders(req *http.Request) []string {
	res := make([]string, 0)
	authorizations := strings.Split(req.Header.Get("Authorization"), ",")
	for _, v := range authorizations {
		if strings.Contains(v, "AdditionalHeaders:") {
			return strings.Split(v[len("AdditionalHeaders:")+1:], ";")
		}
	}
	str := req.URL.Query().Get("x-oss-additional-headers")
	if str != "" {
		return strings.Split(str, ";")
	}
	return res
}

// Copy from OSS SDK (oss/auth.go)
func getAdditionalHeaderKeys(req *http.Request) ([]string, map[string]string) {
	var keysList []string
	keysMap := make(map[string]string)
	srcKeys := make(map[string]string)
	for k := range req.Header {
		srcKeys[strings.ToLower(k)] = ""
	}

	for _, v := range getAdditionalHeaders(req) {
		if _, ok := srcKeys[strings.ToLower(v)]; ok {
			keysMap[strings.ToLower(v)] = ""
		}
	}

	for k := range keysMap {
		keysList = append(keysList, k)
	}
	sort.Strings(keysList)

	return keysList, keysMap
}

// Copy from OSS SDK
func SignRpcRequest(accessKeySecret string, request *http.Request) string {

	stringToSign := buildRpcStringToSign(request)
	signature := Sign(accessKeySecret, stringToSign, "&")
	return signature
}

// Copy from OSS SDK
func Sign(accessKeySecret, stringToSign, secretSuffix string) string {
	secret := accessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

// Copy from OSS SDK
func ShaHmac1(source, secret string) string {
	key := []byte(secret)
	hmac := hmac.New(sha1.New, key)
	hmac.Write([]byte(source))
	signedBytes := hmac.Sum(nil)
	signedString := base64.StdEncoding.EncodeToString(signedBytes)
	return signedString
}

// Copy from OSS SDK
func buildRpcStringToSign(request *http.Request) (stringToSign string) {
	signParams := make(map[string]string)
	for key, value := range request.URL.Query() {
		if key == "Signature" {
			continue
		}
		signParams[key] = value[0]
	}
	stringToSign = GetUrlFormedMap(signParams)
	stringToSign = strings.Replace(stringToSign, "+", "%20", -1)
	stringToSign = strings.Replace(stringToSign, "*", "%2A", -1)
	stringToSign = strings.Replace(stringToSign, "%7E", "~", -1)
	stringToSign = url.QueryEscape(stringToSign)
	stringToSign = request.Method + "&%2F&" + stringToSign
	return
}

// Copy from OSS SDK
func GetUrlFormedMap(source map[string]string) (urlEncoded string) {
	urlEncoder := url.Values{}
	for key, value := range source {
		urlEncoder.Add(key, value)
	}
	urlEncoded = urlEncoder.Encode()
	return
}