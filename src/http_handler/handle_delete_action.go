// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package http_handler

import (
	"net/http"
	"regexp"
	"time"

	"oceanpass/src/common"
	"oceanpass/src/oss_pkg"
	prome "oceanpass/src/prometheus"
	"oceanpass/src/s3_pkg"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"

	"go.uber.org/zap"
)

// TODO: finish http.Method DELETE
func (h *HttpHandler) HandleHttpDelete(w http.ResponseWriter, r *http.Request) {
	// Input
	// ---- DELETE {URL} = DELETE {BucketName}/{ObjectName}?{params}

	// ---- {BucketName}/{ObjectName}
	BucketName := GetBucketName(r)
	ObjectName := GetObjectName(r, BucketName)
	Logger.Info("HandleHttpGet(): uri analyze",
		zap.String("BucketName", BucketName), zap.String("ObjectName", ObjectName))

	// ---- {params}:
	urlParmas := r.URL.Query()
	Logger.Info("HandleHttpDelete()", zap.Any("urlParmas", urlParmas))

	// Handle action
	startAt := time.Now()

	promeArgs := prome.PromeArgs{
		Bucket:      BucketName,
		Object:      ObjectName,
		Method:      "DELETE",
		StartAt:     startAt,
		AccessKeyId: h.Osspkg.AccessKeyId,
		Code:        "200",
		Unique:      make(map[string]interface{}),
	}
	// Object level option
	if BucketName != "" && ObjectName != "" {
		// AbortMultipartUpload
		if common.GetQueryString(urlParmas, "uploadId") != "" {
			pkgRespMap, pkgErr := h.S3pkg.AbortMultipartUpload(
				BucketName, ObjectName, urlParmas)
			promeArgs.Api = "S3pkg.AbortMultipartUpload"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			WriteHeaderWithPkgResMap(w, pkgRespMap,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			Logger.Info("HandleHttpDelete():h.S3pkg.AbortMultipartUpload",
				zap.Any("w.Header()", w.Header()))
			w.WriteHeader(http.StatusNoContent)
			go prome.Prometheus.Do(promeArgs)
			return
		}

		// DeleteObject
		// if objectname contains "/". for example, a/b
		// go sdk: /bucketname/a%2Fb
		// java sdk: /bucketname/a/b
		resMp, pkgErr := h.S3pkg.DeleteObject(BucketName, ObjectName, urlParmas)
		promeArgs.Api = "S3pkg.DeleteObject"
		if pkgErr != nil {
			ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
			WriteErrorToHeader(ocnAwsError, w)
			WriteHeaderStatusCode(ocnAwsError, w)
			WriteErrorToBody(ocnAwsError, w)
			go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
			return
		}
		WriteHeaderWithPkgResMap(w, resMp,
			TypeClientProvider(h.ClientProvider),
			TypeCloudProvider(h.CloudProvider),
		)
		Logger.Info("HandleHttpHead():DeleteObject",
			zap.Any("w.Header()", w.Header()))
		w.WriteHeader(http.StatusNoContent)
		go prome.Prometheus.Do(promeArgs)
		return
	}

	// Bucket level ops
	if BucketName != "" && ObjectName == "" {
		// DeleteBucketCors
		if urlParmas.Has("cors") {
			var resMp map[string]string
			var pkgErr error
			switch h.CloudProvider {
			case common.K_CLOUD_PROVIDER_OSS:
				resMp, pkgErr = h.Osspkg.DeleteBucketCors(BucketName)
				promeArgs.Api = "Osspkg.DeleteBucketCors"
				// if pkgErr == nil, resMp is null
				if pkgErr != nil {
					ocnOssError := oss_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnOssError, w)
					WriteHeaderStatusCode(ocnOssError, w)
					WriteErrorToBody(ocnOssError, w)
					go prome.Prometheus.Do(promeArgs, ocnOssError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
				WriteHeaderWithPkgResMap(w, resMp,
					TypeClientProvider(h.ClientProvider),
					TypeCloudProvider(h.CloudProvider),
				)
			default:
				resMp, pkgErr = h.S3pkg.DeleteBucketCors(BucketName)
				promeArgs.Api = "S3pkg.DeleteBucketCors"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
				WriteHeaderWithPkgResMap(w, resMp,
					TypeClientProvider(h.ClientProvider),
					TypeCloudProvider(h.CloudProvider),
				)
			}
			w.WriteHeader(http.StatusOK)
			go prome.Prometheus.Do(promeArgs)
			return
		}

		// DeleteBucketPolicy
		if urlParmas.Has("policy") {
			switch h.CloudProvider {
			case common.K_CLOUD_PROVIDER_OSS:
				respHeader, pkgErr := h.Osspkg.DeleteBucketPolicy(BucketName)
				promeArgs.Api = "Osspkg.DeleteBucketPolicy"
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
				for resKey, _ := range respHeader {
					w.Header().Set(resKey, respHeader.Get(resKey))
				}
				go prome.Prometheus.Do(promeArgs)
				return
			default:
				pkgErr := h.S3pkg.DeleteBucketPolicy(BucketName)
				promeArgs.Api = "S3pkg.DeleteBucketPolicy"
				if pkgErr != nil {
					ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
					WriteErrorToHeader(ocnAwsError, w)
					WriteHeaderStatusCode(ocnAwsError, w)
					WriteErrorToBody(ocnAwsError, w)
					go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
					return
				}
				go prome.Prometheus.Do(promeArgs)
				return
			}
		}

		// DeleteBucket
		if len(urlParmas) == 0 {
			// TODO: add error handling
			resMp, pkgErr := h.S3pkg.DeleteBucket(BucketName)
			promeArgs.Api = "S3pkg.DeleteBucket"
			if pkgErr != nil {
				ocnAwsError := s3_pkg.HandleErrorReturn(pkgErr)
				WriteErrorToHeader(ocnAwsError, w)
				WriteHeaderStatusCode(ocnAwsError, w)
				WriteErrorToBody(ocnAwsError, w)
				go prome.Prometheus.Do(promeArgs, ocnAwsError.HttpsResponseErrorStatusCode, pkgErr)
				return
			}
			WriteHeaderWithPkgResMap(w, resMp,
				TypeClientProvider(h.ClientProvider),
				TypeCloudProvider(h.CloudProvider),
			)
			w.WriteHeader(http.StatusNoContent)
			go prome.Prometheus.Do(promeArgs)
			return
		}
	}

	errMsg := "Request not supported yet for DELETE."
	HandleErrorRequest(w, h, errMsg)
	return
}
