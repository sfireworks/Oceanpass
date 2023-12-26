// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package http_handler

import (
	"net/http"
	"oceanpass/src/common"
	"oceanpass/src/oss_pkg"
	prome "oceanpass/src/prometheus"
	"oceanpass/src/s3_pkg"
	"regexp"
	"time"

	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"

	"go.uber.org/zap"
)

// TODO: finish http.Method PUT
/*
first, check whether the url has extra parameters,
if yes, check the extra parameters domain by query, then fetch the xml parameters in body
if no, the request is putobject/bucket(distinguish by name format);
*/

func (h *HttpHandler) HandlePut(w http.ResponseWriter, r *http.Request) {
	// Input
	// judge which PUT API via the request line
	// ---- {URL} = {BucketName}/{ObjectName}?{params}
	BucketName := GetBucketName(r)
	ObjectName := GetObjectName(r, BucketName)
	Logger.Info("HandleHttpPut(): uri analyze",
		zap.String("BucketName", BucketName), zap.String("ObjectName", ObjectName))
	w.Header().Set("Content-Type", "application/xml")
	// Things after question mark '?'
	params := r.URL.Query()
	header := r.Header
	copyAliSource := header.Get("x-oss-copy-source")
	copyAmzSource := header.Get("x-amz-copy-source")
	XOssCallback := header.Get("X-Oss-Callback")
	XOssCallbackVar := header.Get("X-Oss-Callback-Var")

	// Handle action
	startAt := time.Now()

	promeArgs := prome.PromeArgs{
		Bucket:      BucketName,
		Object:      ObjectName,
		Method:      "PUT",
		StartAt:     startAt,
		AccessKeyId: h.Osspkg.AccessKeyId,
		Code:        "200",
		Unique:      make(map[string]interface{}),
	}

	///////////////////
	// Bucket level ops
	if BucketName != "" && ObjectName == "" {
		if params.Has("acl") {
			respKv, pkgErr := h.S3pkg.PutBucketAcl(BucketName, r)
			promeArgs.Api = "S3pkg.PutBucketAcl"
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
			go prome.Prometheus.Do(promeArgs)
			return
		}
		if params.Has("logging") {
			respKv, pkgErr := h.S3pkg.PutBucketLogging(BucketName, r)
			promeArgs.Api = "S3pkg.PutBucketLogging"
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
			go prome.Prometheus.Do(promeArgs)
			return
		}
		if params.Has("cors") {
			respKv, pkgErr := h.S3pkg.PutBucketCors(BucketName, r)
			promeArgs.Api = "S3pkg.PutBucketCors"
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
			go prome.Prometheus.Do(promeArgs)
			return
		}
		if params.Has("policy") {
			var respKv map[string]string
			var pkgErr error
			switch h.CloudProvider {
			case common.K_CLOUD_PROVIDER_OSS:
				pkgErr = h.Osspkg.PutBucketPolicy(BucketName, r)
				promeArgs.Api = "Osspkg.PutBucketPolicy"
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
			case common.K_CLOUD_PROVIDER_COS:
				respKv, pkgErr = h.S3pkg.PutBucketPolicy(BucketName, r)
				promeArgs.Api = "S3pkg.PutBucketPolicy"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
			default:
				respKv, pkgErr = h.S3pkg.PutBucketPolicy(BucketName, r)
				promeArgs.Api = "S3pkg.PutBucketPolicy"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
			}
			//TODO: handle_http_return.go::HandleReturn()
			if h.ClientProvider == common.K_CLIENT_PROVIDER_AMAZON {
				w.WriteHeader(http.StatusNoContent)
			} else if h.ClientProvider == common.K_CLIENT_PROVIDER_ALIYUN {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			go prome.Prometheus.Do(promeArgs)
			return
		}

		//Assume this is putbuckets
		respKv, pkgErr := h.S3pkg.CreateBucket(BucketName)
		promeArgs.Api = "S3pkg.CreateBucket"
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
		go prome.Prometheus.Do(promeArgs)
		return
	}

	///////////////////
	// Object level ops
	if BucketName != "" && ObjectName != "" {
		// PutObjectAcl
		if params.Has("acl") {
			respKv, pkgErr := h.S3pkg.PutObjectAcl(
				BucketName, ObjectName, r)
			promeArgs.Api = "S3pkg.PutObjectAcl"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			Logger.Info("HandleHttpPut():PutObjectAcl",
				zap.Any("w.Header()", w.Header()))
			go prome.Prometheus.Do(promeArgs)
			return
		}
		// UploadPart & UploadPartCopy
		if params.Has("partNumber") {
			partNumber := common.GetQueryString(params, "partNumber")
			uploadId := common.GetQueryString(params, "uploadId")
			Logger.Info("HandleHttpPut()",
				zap.String("PartNumber", partNumber),
				zap.String("uploadID", uploadId))
			//header does not contain copy source, normal upload op
			if copyAliSource == "" && copyAmzSource == "" {
				respKv, pkgErr := h.S3pkg.UploadPart(
					BucketName, ObjectName,
					partNumber, uploadId, r.ContentLength, header, r.Body)
				promeArgs.Api = "S3pkg.UploadPart"
				// go h.Record("UploadPart", h.S3pkg.AccessKeyId, BucketName, ObjectName, beginTs)
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
				WriteHeaderWithPkgResMap(w, respKv,
					TypeClientProvider(h.ClientProvider),
					TypeCloudProvider(h.CloudProvider),
				)
				go prome.Prometheus.Do(promeArgs)
				return
			}
			var copySrc string
			if copyAliSource == "" {
				copySrc = copyAmzSource
			} else {
				copySrc = copyAliSource
			}
			//header contains copy source, copy upload op
			//UploadPartCopy return no ETag
			respKv, resp, pkgErr := h.S3pkg.UploadPartCopy(
				BucketName, ObjectName, partNumber, uploadId, copySrc)
			promeArgs.Api = "S3pkg.UploadPartCopy"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				err := WriteErrorToBody(ocnAwsError, w)
				if err != nil {
					return
				}
				WriteHeaderStatusCode(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			WriteHeaderWithPkgResMap(w, respKv,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			ParseResToHttpResp(w, resp, pkgErr)
			go prome.Prometheus.Do(promeArgs)
			return
		}

		// TODO: Is there any other indicator of copyObject ops?
		// copy object handle using same package.
		if copyAliSource != "" || copyAmzSource != "" {
			//header contains copy source, copy put op
			switch h.CloudProvider {
			case common.K_CLOUD_PROVIDER_OSS:
				respKv, pkgResp, pkgErr := h.Osspkg.CopyObject(
					BucketName, ObjectName, r)
				promeArgs.Api = "Osspkg.CopyObject"
				srcBucket, srcObject := oss_pkg.GetCopySourceAttr(r)
				promeArgs.Unique["src_bucket"] = srcBucket
				promeArgs.Unique["src_object"] = srcObject
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
					promeArgs.Unique["src_bucket"] = srcBucket
					promeArgs.Unique["src_object"] = srcObject
					go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
				for resKey, resVal := range respKv {
					w.Header().Set(resKey, resVal)
				}
				ParseOssResToHttpResp(w, pkgResp, pkgErr)
				go prome.Prometheus.Do(promeArgs)
				return
			default:
				respKv, pkgResp, pkgErr := h.S3pkg.CopyObject(
					BucketName, ObjectName,
					copyAliSource+copyAmzSource, header)
				promeArgs.Api = "S3pkg.CopyObject"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					promeArgs.Unique["src_bucket"] = copyAliSource + copyAmzSource
					promeArgs.Unique["src_object"] = copyAliSource + copyAmzSource
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
		}
		// PutObject
		// Fields indicates callback. Only aliyun has this functionality
		var respKv map[string]string
		var pkgResp []byte
		var pkgErr error
		if XOssCallback != "" {
			respKv, pkgResp, pkgErr = h.Osspkg.PutObject(
				BucketName, ObjectName, r.Body,
				XOssCallback, XOssCallbackVar)
			promeArgs.Api = "Osspkg.PutObject"
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
			for resKey, resVal := range respKv {
				w.Header().Set(resKey, resVal)
			}
			ParseOssResToHttpResp(w, pkgResp, pkgErr)
			go prome.Prometheus.Do(promeArgs)
			return
		}
		respKv, pkgErr = h.S3pkg.PutObject(
			BucketName, ObjectName, header,
			r.Body)
		promeArgs.Api = "S3pkg.PutObject"
		// go h.Record("PutObject", h.S3pkg.AccessKeyId, BucketName, ObjectName, beginTs)
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
		go prome.Prometheus.Do(promeArgs)
		return
	}

	errMsg := "Request not supported yet for PUT."
	HandleErrorRequest(w, h, errMsg)
	return
}
