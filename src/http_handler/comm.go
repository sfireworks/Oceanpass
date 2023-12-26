// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package http_handler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net"

	"database/sql"
	"net/http"
	"net/url"
	"oceanpass/src/common"
	"oceanpass/src/config"
	"oceanpass/src/oss_pkg"
	"oceanpass/src/s3_pkg"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type HttpHandler struct {
	S3pkg          s3_pkg.S3Pkg
	Osspkg         oss_pkg.OssPkg
	CloudProvider  string
	OssDb          *sql.DB
	DBConfig       config.DBConfig
	ClientProvider string
}

type Error struct {
	Code                         string
	Message                      string
	HttpsResponseErrorStatusCode string
	RequestId                    string
	HostId                       string
}

func IsIPv4(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ".")
}

func IsIpPort(host string) bool {
	slice := strings.Split(host, ":")
	if len(slice) != 2 {
		return false
	}
	isIP := IsIPv4(slice[0])
	if !isIP {
		return false
	}
	port, _ := strconv.Atoi(slice[1])
	if port > 65535 || port < 1 {
		return false
	}
	return true
}

func GetBucketName(r *http.Request) (bucketName string) {
	host := r.Host
	if isIpPort := IsIpPort(host); !isIpPort {
		slice := strings.Split(host, ".")
		if len(slice) == 0 {
			return ""
		}
		bucketName = slice[0]
		return bucketName
	}

	uri := r.URL.Path
	if len(uri) > 1 {
		bucketName = uri[1:]
		if len(uri) > 2 {
			BucketObjectSlices := strings.Split(uri, "/")
			bucketName = BucketObjectSlices[1]
		}
	} else if uri == "/" {
		host := r.Host
		hostSlice := strings.Split(host, ".")
		if len(hostSlice) == 5 {
			bucketName = hostSlice[0]
		} else {
			bucketName = ""
		}
	}
	return bucketName
}

func GetObjectName(r *http.Request, bucketName string) (objectName string) {
	host := r.Host
	uri := r.URL.Path
	if isIpPort := IsIpPort(host); !isIpPort {
		objectName = uri[1:]
		return objectName
	}

	if len(uri) > len(bucketName)+2 {
		objectName = uri[len(bucketName)+2:]
	}
	return objectName
}

func GetS3SignatureV4(r *http.Request) (signature string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	signature = authorizations[len(authorizations)-1][len("Signature=")+1:]
	return signature
}

func GetOssSignatureV4(r *http.Request) (signature string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	signature = authorizations[len(authorizations)-1][len("Signature="):]
	return signature
}

func GetOssSignatureV1(r *http.Request) (signature string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ":")
	return authorizations[1]
}

func GetOssSignatureV2(r *http.Request) (signature string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	for _, v := range authorizations {
		if strings.Contains(v, "Signature:") {
			return v[len("Signature:")+1:]
		}
	}
	return ""
}

func GetAccessKeyIdV4(r *http.Request) (accessKeyId string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "Credential=")
	fullStr := tmp[index+len("Credential="):]
	accessKeyId = strings.Split(fullStr, "/")[0]
	return accessKeyId
}

func GetAccessKeyIdV1(r *http.Request) (accessKeyId string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ":")
	tmp := authorizations[0]
	return strings.Split(tmp, " ")[1]
}

func GetAccessKeyIdV2(r *http.Request) (accessKeyId string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "AccessKeyId:")
	fullStr := tmp[index+len("AccessKeyId:"):]
	return fullStr
}

func GetAccessKeyIdFromXAmzCredential(r *http.Request) (accessKeyId string) {
	credential := r.URL.Query().Get("X-Amz-Credential")
	res, _ := url.QueryUnescape(credential)
	accessKeyId = strings.Split(res, "/")[0]
	return accessKeyId
}

func GetRegion(r *http.Request) (region string) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "Credential=")
	fullStr := tmp[index+len("Credential="):]
	region = strings.Split(fullStr, "/")[2]
	return region
}

func GetEndpoint(pkg s3_pkg.S3Pkg, urlQuery url.Values) (endpoint string) {
	endpoint = pkg.Endpoint
	if common.GetQueryString(urlQuery, "endpoint") != "" {
		endpoint = common.GetQueryString(urlQuery, "endpoint")
	} else if common.GetQueryString(urlQuery, "endpoint-url") != "" {
		endpoint = common.GetQueryString(urlQuery, "endpoint-url")
	}
	return endpoint
}

func ParseResToHttpResp(w http.ResponseWriter, pkgResp []byte, pkgErr error) (err error) {
	// Resp handle of calling S3pkg function
	if pkgErr != nil {
		res := common.HandleAwsErrorReturn(pkgErr)
		statusCode, err := strconv.Atoi(res.HttpsResponseErrorStatusCode)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("strconv.Atoi", err))
		}
		w.WriteHeader(statusCode)
		resXML, err := xml.MarshalIndent(res, " ", " ")
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("xml.MarshalIndent", err))
		}
		byteErr := []byte(resXML)
		w.Write(byteErr)
		return err
	}
	if pkgResp != nil {
		_, err = w.Write(pkgResp)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("w.Write", err))
		}
	} else if pkgErr == nil && pkgResp == nil {
		errMsg := fmt.Sprintf("%v\n\r", "ParseResToHttpResp():  pkgResp is empty")
		byteErr := []byte(errMsg)
		_, err = w.Write(byteErr)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("w.Write", err))
		}
	} else {
		//TODO: should be very rare case. handle separately in the future.
		errMsg := ""
		byteErr := []byte(errMsg)
		_, err = w.Write(byteErr)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("w.Write", err))
		}
	}
	return err
}

/*
The Oss error message string is like "oss: service returned error: StatusCode=409,
ErrorCode=FileAlreadyExists, ErrorMessage=\"The object you specified already
exists and can not be overwritten.\", RequestId=641D6D20DAC9123233C770C5"

This function is to extract the "StatusCode", "ErrorCode", "ErrorMessage",
and "RequestId" in this message.
*/
func HandleOssErrorReturn(apiErr error) (ocnErr Error) {
	Logger.Info("HandleOssErrorReturn()", zap.Any("apiErr", apiErr.Error()))

	errInfo := strings.Split(apiErr.Error(), ":")
	if len(errInfo) >= 3 {
		log.Println("errMap", errInfo[2])
		errMap := strings.Split(errInfo[2], ",")
		for _, v := range errMap {
			val := strings.Split(v, "=")
			if len(val) == 2 {
				if is, _ := regexp.MatchString("StatusCode", val[0]); is {
					ocnErr.HttpsResponseErrorStatusCode = val[1]
				} else if is, _ := regexp.MatchString("ErrorCode", val[0]); is {
					ocnErr.Code = val[1]
				} else if is, _ := regexp.MatchString("ErrorMessage", val[0]); is {
					ocnErr.Message = val[1]
				} else if is, _ := regexp.MatchString("RequestId", val[0]); is {
					ocnErr.RequestId = val[1]
				} else if is, _ := regexp.MatchString("HostId", val[0]); is {
					ocnErr.HostId = val[1]
				}

			}
		}
	}
	return ocnErr
}

func ParseOssResToHttpResp(w http.ResponseWriter, pkgResp []byte, pkgErr error) (err error) {
	// Resp handle of calling S3pkg function
	if pkgErr != nil {
		res := HandleOssErrorReturn(pkgErr)
		statusCode, err := strconv.Atoi(res.HttpsResponseErrorStatusCode)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("strconv.Atoi", err))
		}
		w.WriteHeader(statusCode)
		resXML, err := xml.MarshalIndent(res, " ", " ")
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("xml.MarshalIndent", err))
		}
		log.Println("=================resXml is ", string(resXML))
		byteErr := []byte(resXML)
		w.Write(byteErr)
		return err
	}
	if pkgResp != nil {
		_, err = w.Write(pkgResp)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("w.Write", err))
		}
	} else {
		//TODO: should be very rare case. handle separately in the future.
		errMsg := ""
		byteErr := []byte(errMsg)
		_, err = w.Write(byteErr)
		if err != nil {
			Logger.Error("ParseResToHttpResp()", zap.Any("w.Write", err))
		}
	}
	return err
}

func ParseRespHeaderToHttpResp(w http.ResponseWriter, mp map[string]string) (err error) {
	for resKey, resVal := range mp {
		if mp[resKey] == resVal {
			w.Header().Set(resKey, resVal)
		} else {
			errMsg := fmt.Sprintf("[ERROR] for-range(respMp): mp[resKey]:(%s) != resVal(%s). ",
				mp[resKey], resVal)
			Logger.Error("ParseRespHeaderToHttpResp()", zap.Any("w.Write", errMsg))
			byteErr := []byte(errMsg)
			_, err = w.Write(byteErr)
			if err != nil {
				Logger.Error("ParseRespHeaderToHttpResp()", zap.Any("w.Write", err))
			}
		}
	}
	return nil
}

func GetServiceName(r *http.Request) string {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "Credential=")
	fullStr := tmp[index+len("Credential="):]
	return strings.Split(fullStr, "/")[3]
}

func IsStsFromHttp(r *http.Request) bool {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "Credential=")
	fullStr := tmp[index+len("Credential="):]
	sts := strings.Split(fullStr, "/")[3]
	return sts == "sts"
}

/*
Authorization format from different http.Header:

	aws:
	Authorization:[AWS4-HMAC-SHA256 Credential=FUGJYBFOMTYNNRCMGDRH/20230227/ap-shanghai/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=c4912590b450ff87494ca323471a768db9af9729dd75ac3d9b59a4ce0cbedd30]

	oss:
	Authorization:[OSS LTAI5t88XupUCBSretjPz8Wu:Ky69l92rLZZttzCSO0DXXM235UY=]
*/
func GetSignatureParamsFromRequestHeader(r *http.Request) (authorization *common.Authorization, err error) {
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	if len(authorizations) == 0 {
		return nil, fmt.Errorf("[Common] GetSignatureParams: No Authorization")
	}
	authorizations0 := authorizations[0]
	authorizations0s := strings.Split(authorizations0, " ")
	if len(authorizations0s) != 2 {
		return nil, fmt.Errorf("[Common] GetSignatureParams: authorizations0.length() !=2")
	}

	authSignMethod := authorizations0s[0]
	authSignParams0 := authorizations0s[1]

	authorization = new(common.Authorization)
	authorization.SignMethod = authSignMethod
	if authorization.SignMethod == common.K_SIGN_METHOD_OSS {
		authSignParams0s := strings.Split(authSignParams0, ":")
		if len(authSignParams0s) > 1 {
			authorization.AccessKeyId = authSignParams0s[0]
			authorization.Signature = authSignParams0s[1]
		} else {
			Logger.Error("OSS authorization param err", zap.Any("authorizations", authorizations))
			return authorization, nil
		}
		// todo:
		// support oss cli/sdk
		authorization.Signature = authSignParams0s[1]
		return authorization, nil
		//		Logger.Error("[OSS] authorization not support", zap.Any("authorizations", authorizations))
		// return authorization, nil
	} else if authorization.SignMethod == common.K_SIGN_METHOD_AWS4 {
		if len(authorizations) < 3 {
			return nil, fmt.Errorf("[AWS4] GetSignatureParams:No Authorization")
		}
		re, _ := regexp.Compile("([a-z]*)Credential=([a-z]*)")
		loc := re.FindStringIndex(authSignParams0)
		if len(loc) > 1 {
			// authorizations[0]
			credential := authSignParams0[loc[1]:]
			credentials := strings.Split(credential, "/")
			if len(credentials) < 5 {
				Logger.Error("AWS4-HMAC-SHA256 authorization param err",
					zap.Any("credential", credential), zap.Any("authorizations", authorizations))
				return authorization, fmt.Errorf("[AWS4] GetSignatureParams: authorization param err")
			}

			// authorizations[1:]
			SignedHeaders := strings.Split(authorizations[1], "=")
			Signature := strings.Split(authorizations[2], "=")
			if len(SignedHeaders) < 2 || len(Signature) < 2 {
				Logger.Error("AWS4-HMAC-SHA256 authorization param err",
					zap.Any("SignedHeaders", SignedHeaders), zap.Any("Signature", Signature))
				return authorization, fmt.Errorf("[AWS4] GetSignatureParams: authorization param err")
			}

			// set output
			authorization.AwsCredential = credential
			authorization.AccessKeyId = credentials[0]
			authorization.AwsAuthDate = credentials[1]
			authorization.AwsAuthRegion = credentials[2]
			authorization.AwsAuthServiceName = credentials[3]
			authorization.AwsAuthRequestName = credentials[4]

			authorization.AwsSignedHeaders = SignedHeaders[1]
			authorization.Signature = Signature[1]
		}
	} else {
		Logger.Error("[xxx] authorization not support")
		return authorization, fmt.Errorf("[xxx] GetSignatureParams: authorization param err")
	}
	Logger.Info("authorization: get params ok.",
		zap.Any("Input(authorizations)", authorizations),
		zap.Any("Output(authorization)", authorization))
	return authorization, nil
}

func GetCredentialWithoutAccessKey(req *http.Request) string {
	authorizations := strings.Split(req.Header.Get("Authorization"), ",")
	tmp := authorizations[0]
	index := strings.Index(tmp, "Credential=")
	fullStr := tmp[index+len("Credential="):]
	arr := strings.Split(fullStr, "/")[1:]
	return strings.Join(arr, "/")
}

func GetOssSignatureVersion(req *http.Request) string {
	authorizations := strings.Split(req.Header.Get("Authorization"), " ")
	return authorizations[0]
}

func GetOssStsSignatureVersion(req *http.Request) string {
	str := req.URL.Query().Get("SignatureVersion")
	if str == "1.0" {
		return "OSS"
	} else if str == "2.0" {
		return "OSS2"
	} else {
		return "OSS4"
	}
}

func GetStsToken(r *http.Request) string {
	res := r.Header.Get("X-Amz-Security-Token")
	if res == "" {
		res = r.Header.Get("X-Oss-Security-Token")
	}
	return res
}

func GetAssumeRoleParams(strBody string) map[string]string {
	strBody, _ = url.QueryUnescape(strBody)
	arr := strings.Split(strBody, "&")
	res := make(map[string]string, 0)

	for i := 0; i < len(arr); i++ {
		tmp := strings.Split(arr[i], "=")
		res[tmp[0]] = tmp[1]
	}
	return res
}

type RequestResp struct {
	ErrCode int32  `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

type RequestPart struct {
	PartNumber string `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type RequestBodyXML struct {
	Name  string        `xml:"CompleteMultipartUpload"`
	Parts []RequestPart `xml:"Part"`
}

// generate json format response string
func RespJsonGen(errCode int32, errMsg string) (string, error) {
	resp := RequestResp{
		ErrCode: errCode,
		ErrMsg:  errMsg,
	}

	jsonResp, err := json.Marshal(resp)
	return string(jsonResp), err
}

func HandleErrorRequest(w http.ResponseWriter, h *HttpHandler, errMsg string) {
	Logger.Error(errMsg)
	HttpsStatusCode := strconv.Itoa(http.StatusNotImplemented)
	ocnOssError := oss_pkg.Error{
		Code:                         "UnsupportedRequest",
		RequestId:                    "No RequestId due to request error.",
		Message:                      errMsg,
		HttpsResponseErrorStatusCode: HttpsStatusCode,
	}
	if ocnOssError.HostId == "" {
		re, _ := regexp.Compile("([^://]+)$")
		endpoint := re.FindString(h.Osspkg.Endpoint)
		ocnOssError.HostId = endpoint
	}
	WriteErrorToHeader(ocnOssError, w)
	WriteHeaderStatusCode(ocnOssError, w)
	WriteErrorToBody(ocnOssError, w)
}

// Write Header
type optionType string

type (
	optionValue struct {
		Value interface{}
		Type  optionType
	}
	Option func(map[string]optionValue) error
)

const (
	optionParam optionType = "OptionParameter"
)

func setHeader(key string, value interface{}) Option {
	return func(params map[string]optionValue) error {
		if value == nil {
			return nil
		}
		params[key] = optionValue{value, optionParam}
		return nil
	}
}

// function: Get params from []Option, and return these params with a map.
// input: []Option
// output: map[string]interface{}
func GetParams(options []Option) (map[string]interface{}, error) {
	// Option
	params := map[string]optionValue{}
	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return nil, err
			}
		}
	}

	outputParams := map[string]interface{}{}
	// Serialize
	for k, v := range params {
		if v.Type == optionParam {
			vs := params[k]
			outputParams[k] = vs.Value.(string)
		}
	}
	return outputParams, nil
}

// Business logic
type TypeClientProvider string
type TypeCloudProvider string

// local variable for business logic function
const (
	keyClientProvider optionType = "ClientProvider"
	keyCloudProvider  optionType = "CloudProvider"
)

// Transter []interface{} to []Option
func TransterOptions(inputOptions []interface{}) []Option {
	var options []Option
	for _, v := range inputOptions {
		var key string
		var value interface{}
		switch vv := v.(type) {
		case TypeClientProvider:
			key = string(keyClientProvider)
			value = string(vv)
		case TypeCloudProvider:
			key = string(keyCloudProvider)
			value = string(vv)
		default:
			key = string(optionParam)
			value = vv
		}
		opt := setHeader(key, value)
		options = append(options, opt)
	}
	return options
}

// Business logic
func WriteHeaderWithPkgResMap(w http.ResponseWriter, headerMap map[string]string,
	options ...interface{}) error {
	// 1.Get params
	transteredOptions := TransterOptions(options)
	params, _ := GetParams(transteredOptions)

	// 2.Business logic
	// 2.1 get two params
	clientProvider := ""
	cloudProvider := ""
	if _, ok := params["ClientProvider"]; ok {
		clientProvider = params["ClientProvider"].(string)
	}
	if _, ok := params["CloudProvider"]; ok {
		cloudProvider = params["CloudProvider"].(string)
	}

	// 2.2 set Header
	for resKey, resVal := range headerMap {
		resKey = strings.ToLower(resKey)
		re, err := regexp.Compile(common.K_RESPONSE_HEADER_LOWER_CASE_AMAZON)
		if err != nil {
			log.Fatalln("regexp Compile failed!")
		}
		matchArr := re.FindStringSubmatch(resKey)
		if matchArr != nil {
			if cloudProvider == common.K_CLOUD_PROVIDER_OSS &&
				clientProvider == common.K_CLIENT_PROVIDER_ALIYUN {
				newKeyBytes := re.ReplaceAll([]byte(resKey),
					[]byte(common.K_RESPONSE_HEADER_LOWER_CASE_ALIYUN))
				w.Header().Set(string(newKeyBytes), resVal)
				continue
			}
			if cloudProvider == common.K_CLOUD_PROVIDER_COS {
				// TODO: test it when endpoint is cos in future.
				newKeyBytes := re.ReplaceAll([]byte(resKey),
					[]byte(common.K_RESPONSE_HEADER_LOWER_CASE_TENCENT))
				w.Header().Set(string(newKeyBytes), resVal)
				continue
			}
		}
		w.Header().Set(resKey, resVal)
	}
	return nil
}
