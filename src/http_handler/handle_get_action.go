// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package http_handler

import (
	"io"
	"net/http"
	"oceanpass/src/common"
	"oceanpass/src/oss_pkg"
	prome "oceanpass/src/prometheus"
	"oceanpass/src/s3_pkg"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"regexp"
	"time"

	"go.uber.org/zap"
)

func (h *HttpHandler) HandleHttpGet(w http.ResponseWriter, r *http.Request) {
	// Input
	// ---- GET {URL} = GET {BucketName}/{ObjectName}?{params}

	// ---- {BucketName}/{ObjectName}
	BucketName := GetBucketName(r)
	ObjectName := GetObjectName(r, BucketName)
	Logger.Info("HandleHttpGet(): uri analyze",
		zap.String("BucketName", BucketName), zap.String("ObjectName", ObjectName))

	// ---- {params}
	urlParams := r.URL.Query()
	// defalut endpoint is from S3config (S3config is todo work)

	// Output
	var pkgResp []byte
	var ioResp io.ReadCloser
	var pkgErr error
	w.Header().Set("Content-Type", "application/xml")

	// Handle action
	startAt := time.Now()

	promeArgs := prome.PromeArgs{
		Bucket:      BucketName,
		Object:      ObjectName,
		Method:      "GET",
		StartAt:     startAt,
		AccessKeyId: h.Osspkg.AccessKeyId,
		Code:        "200",
		Unique:      make(map[string]interface{}),
	}

	// Object level option
	if BucketName != "" && ObjectName != "" {
		//GetObjectMeta
		if urlParams.Has("objectMeta") {
			//Perhaps Aliyun sdk GetObjectMeta is more lightweight
			pkgResp, pkgErr := h.Osspkg.GetObjectMeta(
				BucketName, ObjectName, urlParams.Get("versionId"))
			promeArgs.Api = "Osspkg.GetObjectMeta"
			if pkgErr != nil {
				//TODO : no errcode
				//default: 404 Not found
				ocnOssError := oss_pkg.HandleErrorReturn(pkgErr)
				if ocnOssError.HostId == "" {
					re, _ := regexp.Compile("([^://]+)$")
					endpoint := re.FindString(h.Osspkg.Endpoint)
					ocnOssError.HostId = BucketName + "." + endpoint
				}
				WriteErrorToHeader(ocnOssError, w)
				WriteHeaderStatusCode(ocnOssError, w)
				WriteErrorToBody(ocnOssError, w)
				Logger.Error("HandleHttpHead()", zap.Any("h.S3pkg.GetObjectMeta", pkgErr))
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			for key, _ := range pkgResp {
				w.Header().Set(key, pkgResp.Get(key))
			}
			go prome.Prometheus.Do(promeArgs)
			return
		}
		// GetObjectAcl
		if urlParams.Has("acl") {
			var pkgResp []byte
			var respKv map[string]string
			var pkgErr error
			switch h.ClientProvider {
			case common.K_CLIENT_PROVIDER_ALIYUN:
				pkgResp, respKv, pkgErr = h.Osspkg.GetObjectACL(
					BucketName, ObjectName, urlParams)
				promeArgs.Api = "Osspkg.GetObjectACL"
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
			default:
				pkgResp, respKv, pkgErr = h.S3pkg.GetObjectAcl(
					BucketName, ObjectName, urlParams)
				promeArgs.Api = "S3pkg.GetObjectACL"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			ParseResToHttpResp(w, pkgResp, pkgErr)
			go prome.Prometheus.Do(promeArgs)
			return
		}
		// ListParts
		if common.GetQueryString(urlParams, "uploadId") != "" {
			pkgResp, respKv, pkgErr := h.S3pkg.ListParts(
				BucketName, ObjectName, urlParams)
			promeArgs.Api = "S3pkg.ListParts"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.ListParts", zap.Any("error", pkgErr))
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
		// ListMultipartUploads
		if urlParams.Has("uploads") {
			pkgResp, respKv, pkgErr := h.S3pkg.ListMultipartUploads(
				BucketName, ObjectName, urlParams)
			promeArgs.Api = "S3pkg.ListMultipartUploads"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.ListMultipartUploads", zap.Any("error", pkgErr))
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
		// GetObject
		// if urlParams.Has(("partNumber")) || len(urlParams) == 0 ||
		// 	common.GetQueryString(urlParams, "versionId") != "" ||
		// 	(common.GetQueryString(urlParams, "response-cache-control") != "" &&
		// 		common.GetQueryString(urlParams, "response-content-disposition") != "" &&
		// 		common.GetQueryString(urlParams, "response-content-encoding") != "" &&
		// 		common.GetQueryString(urlParams, "response-content-language") != "" &&
		// 		common.GetQueryString(urlParams, "response-expires") != "") {
		var pkgRespHeader map[string]string
		var pkgRespBody io.ReadCloser
		pkgRespHeader, pkgRespBody, pkgErr = h.S3pkg.GetObject(
			BucketName, ObjectName,
			urlParams, r.Header)
		promeArgs.Api = "S3pkg.GetObject"
		if pkgRespBody != nil {
			defer pkgRespBody.Close()
		}
		if pkgErr != nil {
			ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
			WriteErrorToHeader(ocnAwsError, w)
			WriteHeaderStatusCode(ocnAwsError, w)
			WriteErrorToBody(ocnAwsError, w)
			Logger.Error("HandleHttpGet():h.S3pkg.GetObject", zap.Any("error", pkgErr))
			go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
			return
		} else {
			ioResp = pkgRespBody
			WriteHeaderWithPkgResMap(w, pkgRespHeader,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
		}
		if ioResp != nil {
			io.Copy(w, ioResp)
		} else {
			ParseResToHttpResp(w, pkgResp, pkgErr)
		}
		go prome.Prometheus.Do(promeArgs)
		return
		//}
	}
	// Bucket level ops
	// all bucket ops except ListBuckets.
	// but has ListObjectsV2, ListObjects
	if BucketName != "" && ObjectName == "" {
		// GetBucketInfo
		if urlParams.Has("bucketInfo") {
			pkgResp, respKv, pkgErr := h.Osspkg.GetBucketInfo(
				BucketName, urlParams)
			promeArgs.Api = "Osspkg.GetBucketInfo"
			if pkgErr != nil {
				ocnOssError := oss_pkg.HandleErrorReturn(pkgErr)
				if ocnOssError.HostId == "" {
					re, _ := regexp.Compile("([^://]+)$")
					endpoint := re.FindString(h.Osspkg.Endpoint)
					ocnOssError.HostId = BucketName + "." + endpoint
				}
				WriteHeaderStatusCode(ocnOssError, w)
				WriteErrorToBody(ocnOssError, w)
				Logger.Error("HandleHttpGet():h.Osspkg.GetBucketInfo", zap.Any("error", pkgErr))
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
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
		// GetBucketAcl
		if urlParams.Has("acl") {
			var pkgResp []byte
			var respKv map[string]string
			var pkgErr error
			switch h.CloudProvider {
			case common.K_CLOUD_PROVIDER_OSS:
				pkgResp, respKv, pkgErr = h.Osspkg.GetBucketACL(BucketName)
				promeArgs.Api = "Osspkg.GetBucketACL"
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
			default:
				pkgResp, respKv, pkgErr = h.S3pkg.GetBucketAcl(BucketName)
				promeArgs.Api = "S3pkg.GetBucketACL"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			ParseResToHttpResp(w, pkgResp, pkgErr)
			go prome.Prometheus.Do(promeArgs)
			return
		}
		// GetBucketLocation
		if urlParams.Has("location") {
			pkgResp, respKv, pkgErr := h.S3pkg.GetBucketLocation(
				BucketName, urlParams)
			promeArgs.Api = "S3pkg.GetBucketLocation"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.GetBucketLocation", zap.Any("error", pkgErr))
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
		// GetBucketCors
		if urlParams.Has("cors") {
			pkgResp, respKv, pkgErr := h.S3pkg.GetBucketCors(
				BucketName, urlParams)
			promeArgs.Api = "S3pkg.GetBucketCors"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.GetBucketCors", zap.Any("error", pkgErr))
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
		// GetBucketPolicyStatus
		if urlParams.Has("policyStatus") {
			pkgResp, respKv, pkgErr := h.S3pkg.GetBucketPolicyStatus(
				BucketName, urlParams)
			promeArgs.Api = "S3pkg.GetBucketPolicyStatus"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.GetBucketPolicyStatus", zap.Any("error", pkgErr))
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
		// GetBucketPolicy
		if urlParams.Has("policy") {
			pkgResp, respKv, pkgErr := h.Osspkg.GetBucketPolicy(BucketName)
			promeArgs.Api = "Osspkg.GetBucketPolicy"
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
				Logger.Error("HandleHttpGet():h.Osspkg.GetBucketPolicy", zap.Any("error", pkgErr))
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
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
		// GetBucketLogging
		if urlParams.Has("logging") {
			pkgResp, respKv, pkgErr := h.S3pkg.GetBucketLogging(
				BucketName, urlParams)
			promeArgs.Api = "S3pkg.GetBucketLogging"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.GetBucketLogging", zap.Any("error", pkgErr))
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
		// ListObjectsV2
		if common.GetQueryString(urlParams, "list-type") == "2" {
			pkgResp, respKv, pkgErr := h.S3pkg.ListObjectsV2(
				BucketName, urlParams)
			promeArgs.Api = "S3pkg.ListObjectsV2"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				Logger.Error("HandleHttpGet():h.S3pkg.ListObjectsV2", zap.Any("error", pkgErr))
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

		// ListObjects
		// List all objects under certain bucket.
		// awscli has "url", but aws s3 protocol does not have it.
		// TODO: enum all cli/sdk to test this case.
		// // if urlParams.Has("url") {
		pkgResp, respKv, pkgErr := h.S3pkg.ListObjects(
			BucketName, urlParams)
		promeArgs.Api = "S3pkg.ListObjects"
		if pkgErr != nil {
			ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
			WriteErrorToHeader(ocnAwsError, w)
			WriteHeaderStatusCode(ocnAwsError, w)
			WriteErrorToBody(ocnAwsError, w)
			Logger.Error("HandleHttpGet():h.S3pkg.ListObjects", zap.Any("error", pkgErr))
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
	// Account level ops: ListBuckets
	if BucketName == "" && ObjectName == "" {
		// ListBuckets
		if len(urlParams) == 0 && r.RequestURI == "/" {
			mp := make(map[string]string)
			if common.GetQueryString(urlParams, "max-keys") == "" {
				mp["max-keys"] = "1000"
			}
			pkgResp, respKv, pkgErr := h.Osspkg.ListBuckets(BucketName, mp)
			promeArgs.Api = "Osspkg.ListBuckets"
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
				Logger.Error("HandleHttpGet():h.Osspkg.ListBuckets", zap.Any("error", pkgErr))
				go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
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
	}

	errMsg := "Request not supported yet for GET."
	HandleErrorRequest(w, h, errMsg)
	return
}
