// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package http_handler

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"oceanpass/src/dbops"
	"oceanpass/src/oss_pkg"
	prome "oceanpass/src/prometheus"
	"oceanpass/src/s3_pkg"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"regexp"
	"strconv"
	"strings"
	"time"

	sts "github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"go.uber.org/zap"
)

// TODO: do not code POST request up to now.
func (h *HttpHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	// Input
	// judge which POST API via the request line
	// ---- {URL} = {BucketName}/{ObjectName}?{params}

	// ---- {BucketName}/{ObjectName}
	// get BucketName and ObjectName
	BucketName := GetBucketName(r)
	ObjectName := GetObjectName(r, BucketName)
	Logger.Info("HandleHttpPost(): uri analyze",
		zap.String("BucketName", BucketName), zap.String("ObjectName", ObjectName))
	uri := r.URL.Path
	reqURI := r.RequestURI
	// ---- {params}: len(reqURISlice) == 2
	params := r.URL.Query()
	header := r.Header
	XOssCallback := header.Get("X-Oss-Callback")
	XOssCallbackVar := header.Get("X-Oss-Callback-Var")
	// Output
	bodyBytes, err := ioutil.ReadAll(r.Body)
	userAgent := r.Header.Get("User-Agent")
	if err != nil {
		Logger.Error("HandleHttpPost(): read request body failed")
	}
	// reuse r.body
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	Logger.Info("HandleHttpPost()",
		zap.String("uri", uri),
		zap.String("requestURI", reqURI),
		zap.Any("params", params))
	Logger.Info("HandleHttpPost()",
		zap.Any("request Body", bodyBytes))

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		Logger.Error("HandleHttpPost(): dumprequest failed")
	}
	Logger.Info("HandleHttpPost()",
		zap.Any("request all", string(requestDump)))

	// Handle action
	startAt := time.Now()

	promeArgs := prome.PromeArgs{
		Bucket:      BucketName,
		Object:      ObjectName,
		Method:      "POST",
		StartAt:     startAt,
		AccessKeyId: h.Osspkg.AccessKeyId,
		Code:        "200",
		Unique:      make(map[string]interface{}),
	}
	///////////////////
	// User level ops
	if BucketName == "" && ObjectName == "" {
		//AssumeRole
		var pkgResp []byte
		var pkgErr error
		var code int
		if strings.Contains(userAgent, "aws") && GetServiceName(r) == "sts" {
			bodyBytes, _ := ioutil.ReadAll(r.Body)
			strBody := string(bodyBytes)
			resMap := GetAssumeRoleParams(strBody)
			pkgResp, code, pkgErr = h.Osspkg.AssumeRole(resMap["RoleArn"], resMap["RoleSessionName"], resMap["DurationSeconds"])
			promeArgs.Api = "Osspkg.AssumeRolAwsSts"
			if pkgErr != nil {
				Logger.Error("HandleHttpPost():AssumeRole", zap.Any("h.Osspkg.AssumeRole", err))
				ocnOssError := oss_pkg.HandleErrorReturn(pkgErr)
				if ocnOssError.HostId == "" {
					re, _ := regexp.Compile("([^://]+)$")
					endpoint := re.FindString(h.Osspkg.Endpoint)
					ocnOssError.HostId = BucketName + "." + endpoint
				}
				WriteErrorToHeader(ocnOssError, w)
				WriteHeaderStatusCode(ocnOssError, w)
				WriteErrorToBody(ocnOssError, w)
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			ParseResToHttpResp(w, pkgResp, pkgErr)
			err := h.PersistAssumeRoleResp(pkgResp)
			if err != nil {
				Logger.Error("HTTPHandler():http.MethodPost", zap.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			go prome.Prometheus.Do(promeArgs)
			return
		}
		if r.Header.Get("X-Acs-Action") == "AssumeRole" {
			roleArn := r.URL.Query().Get("RoleArn")
			roleArn, _ = url.QueryUnescape(roleArn)
			sessionName := r.URL.Query().Get("RoleSessionName")
			sessionName, _ = url.QueryUnescape(sessionName)
			durationSeconds := r.URL.Query().Get("DurationSeconds")
			durationSeconds, _ = url.QueryUnescape(durationSeconds)
			pkgResp, code, pkgErr = h.Osspkg.AssumeRole(roleArn, sessionName, durationSeconds)
			if pkgErr != nil {
				Logger.Error("HandleHttpPost():AssumeRole", zap.Any("h.Osspkg.AssumeRole", err))
				ocnOssError := oss_pkg.HandleAssumeRoleErrorReturn(pkgErr, code)
				WriteErrorToHeader(ocnOssError, w)
				WriteHeaderStatusCode(ocnOssError, w)
				WriteErrorToBody(ocnOssError, w)
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			ParseResToHttpResp(w, pkgResp, pkgErr)
			err := h.PersistAssumeRoleResp(pkgResp)
			if err != nil {
				Logger.Error("HTTPHandler():http.MethodPost", zap.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			go prome.Prometheus.Do(promeArgs)
			return
		}
	}
	///////////////////
	// Bucket level ops
	if BucketName != "" && ObjectName == "" {
		//DeleteObjects
		if params.Has("delete") {
			bodyBytes, _ := ioutil.ReadAll(r.Body)
			//TODO: Groom the code
			pkgResp, respKv, pkgErr := h.S3pkg.DeleteObjects(BucketName, bodyBytes)
			promeArgs.Api = "S3pkg.DeleteObjects"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			ParseResToHttpResp(w, pkgResp, pkgErr)
			go prome.Prometheus.Do(promeArgs)
			return
		}
		//ListObjects
		Logger.Info("HandleHttpPost():origin")
		var jsonResp string
		jsonResp = "{}"

		// parse json
		len := r.ContentLength
		body := make([]byte, len)
		r.Body.Read(body)

		var Req s3_pkg.ListObjectsReq
		err = json.Unmarshal(body, &Req)

		Logger.Info("HandleHttpPost()",
			zap.Any("r.Body", r.Body),
			zap.String("body", string(body)),
			zap.Any("Req", Req))

		if err != nil {
			Logger.Error("HandleHttpPost()", zap.Any("Unmarshal err", err))
			jsonResp, err = RespJsonGen(-1, "Invalid json format")
			fmt.Fprintf(w, jsonResp)
			return
		}

		err = s3_pkg.CheckListObjectsParams(Req)
		if err != nil {
			Logger.Error("HandleHttpPost()", zap.Any("CheckListObjectsParams err", err))
			jsonResp, err = RespJsonGen(-2, "Invalid HTTP request parameters")
			fmt.Fprintf(w, jsonResp)
			return
		}

		mp := make(map[string]string)
		mp["delimiter"] = Req.Delimiter
		mp["marker"] = Req.Marker
		mp["max-keys"] = strconv.Itoa(int(Req.Maxkeys))
		mp["prefix"] = Req.Prefix
		// handle request
		pkgResp, respKv, pkgErr := h.Osspkg.ListObjects(BucketName, mp)
		promeArgs.Api = "Osspkg.ListObjects"
		WriteHeaderWithPkgResMap(w, respKv,
			TypeClientProvider(h.ClientProvider),
			TypeCloudProvider(h.CloudProvider),
		)
		ParseResToHttpResp(w, pkgResp, pkgErr)
		go prome.Prometheus.Do(promeArgs)
		return
	}
	///////////////////
	// Object level ops
	if BucketName != "" && ObjectName != "" {
		// CreateMultipartUpload()
		//TODO to check why uploadVal is an array and use uploadVal[0]
		if uploadVal, ok := params["uploads"]; ok && uploadVal[0] == "" {
			Logger.Info("HandleHttpPost():CreateMultipartUpload",
				zap.String("urlQuery uploads", params["uploads"][0]),
				zap.Int("urlQuery uploads length", len(params["uploads"])))
			pkgResp, respKv, pkgErr := h.S3pkg.CreateMultipartUpload(BucketName, ObjectName, params, header)
			promeArgs.Api = "S3pkg.CreateMultipartUpload"
			// go h.Record("CreateMultipartUpload", h.S3pkg.AccessKeyId, BucketName, ObjectName, beginTs)
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			ParseResToHttpResp(w, pkgResp, pkgErr)
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			go prome.Prometheus.Do(promeArgs)
			return
		}
		//CompleteMultipartUpload
		//TODO to check why uploadIdVal is an array and use uploadIdVal[0]
		if uploadIdVal, ok := params["uploadId"]; ok && uploadIdVal[0] != "" {
			Logger.Info("HandleHttpPost():CompleteMultipartUpload",
				zap.String("urlQuery uploadId", params["uploadId"][0]),
				zap.Int("urlQuery uploadId length", len(params["uploadId"])))

			var requestParts RequestBodyXML
			err := xml.Unmarshal([]byte(bodyBytes), &requestParts)
			if err != nil {
				Logger.Error("HandleHttpPost(): xml unmarshal failed")
			}
			Logger.Info("HandleHttpPost()",
				zap.Any("requestParts", requestParts.Parts))

			partNumEtag := make(map[int]string)
			for _, part := range requestParts.Parts {
				partNum, _ := strconv.Atoi(part.PartNumber)
				partEtag := part.ETag
				partNumEtag[partNum] = partEtag
			}

			var pkgResp []byte
			var pkgErr error
			var respKv map[string]string
			// if XOssCallback != ""  {
			// TODO: aliyun_go_sdk do not have XOssCallbackVar, may check XOssCallback and test.
			if XOssCallback != "" && XOssCallbackVar != "" {
				pkgResp, pkgErr = h.Osspkg.CompleteMultipartUpload(
					BucketName, ObjectName, params, partNumEtag,
					XOssCallback, XOssCallbackVar)
				promeArgs.Api = "Osspkg.CompleteMultipartUpload"
				if pkgErr != nil {
					ocnOssError := oss_pkg.HandleErrorReturn(pkgErr)
					if ocnOssError.HostId == "" {
						re, _ := regexp.Compile("([^://]+)$")
						endpoint := re.FindString(h.Osspkg.Endpoint)
						ocnOssError.HostId = BucketName + "." + endpoint
					}
					WriteErrorToHeader(ocnOssError, w)
					WriteHeaderStatusCode(ocnOssError, w)
					WriteErrorToBody(ocnOssError, w)
					go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
			} else {
				pkgResp, respKv, pkgErr = h.S3pkg.CompleteMultipartUpload(
					BucketName, ObjectName, params, partNumEtag)
				promeArgs.Api = "S3pkg.CompleteMultipartUpload"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}

			}
			ParseResToHttpResp(w, pkgResp, pkgErr)
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			go prome.Prometheus.Do(promeArgs)
			return
		}
	}

	Logger.Error("Request not supported yet for HEAD")
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func (h *HttpHandler) PersistAssumeRoleResp(pkgResp []byte) error {
	response := sts.AssumeRoleResponse{}
	err := xml.Unmarshal(pkgResp, &response)
	if err != nil {
		Logger.Error("HTTPHandler():http.MethodPost", zap.Any("err", err))
		return err
	}
	err = dbops.InsertIntoDB(
		response.Credentials,
		h.S3pkg.Endpoint,
		h.OssDb,
		h.DBConfig.DbBase.TableName)
	if err != nil {
		Logger.Error("HTTPHandler():http.MethodPost", zap.Any("err", err))
		return err
	}
	return nil
}
