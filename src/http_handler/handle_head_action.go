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

func (h *HttpHandler) HandleHttpHead(w http.ResponseWriter, r *http.Request) {
	// Input
	// ---- HEAD {URL} = HEAD {BucketName}/{ObjectName}?{params}

	// ---- {BucketName}/{ObjectName}
	BucketName := GetBucketName(r)
	ObjectName := GetObjectName(r, BucketName)
	Logger.Info("HandleHttpHead(): uri analyze",
		zap.String("BucketName", BucketName), zap.String("ObjectName", ObjectName))

	// ---- {params}
	urlParmas := r.URL.Query()
	// defalut endpoint is from S3config (S3config is todo work)

	// Handle action
	startAt := time.Now()

	promeArgs := prome.PromeArgs{
		Bucket:      BucketName,
		Object:      ObjectName,
		Method:      "HEAD",
		StartAt:     startAt,
		AccessKeyId: h.Osspkg.AccessKeyId,
		Code:        "200",
		Unique:      make(map[string]interface{}),
	}
	// Object level option
	if BucketName != "" && ObjectName != "" {
		// Aliyun.GetObjectMeta
		// Perhaps Aliyun sdk GetObjectMeta is more lightweight
		if urlParmas.Has("objectMeta") &&
			h.ClientProvider == common.K_CLIENT_PROVIDER_ALIYUN {
			pkgResp, pkgErr := h.Osspkg.GetObjectMeta(
				BucketName, ObjectName, urlParmas.Get("versionId"))
			promeArgs.Api = "Osspkg.GetObjectMeta"
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
			for key, _ := range pkgResp {
				w.Header().Set(key, pkgResp.Get(key))
			}
			go prome.Prometheus.Do(promeArgs)
			return
		}

		// Aliyun.GetObjectDetailedMeta
		// Theoretically Aliyun.GetObjectDetailedMeta == aws.HeadObject
		// TODO: waiting for aliyun's response.
		if /*len(urlParmas) == 0 && */
		urlParmas.Has("objectDetailedMeta") &&
			h.ClientProvider == common.K_CLIENT_PROVIDER_ALIYUN {
			// TODO: test performance
			pkgResp, pkgErr := h.Osspkg.GetObjectDetailedMeta(
				BucketName, ObjectName)
			promeArgs.Api = "Osspkg.GetObjectDetailedMeta"
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
			for key, _ := range pkgResp {
				w.Header().Set(key, pkgResp.Get(key))
			}
			go prome.Prometheus.Do(promeArgs)
			return
		}

		// HeadObject
		resMp, pkgErr := h.S3pkg.HeadObject(
			BucketName, ObjectName, urlParmas)
		promeArgs.Api = "S3pkg.HeadObject"
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

		Logger.Info("HandleHttpHead()", zap.Any("w.Header()", w.Header()))
		go prome.Prometheus.Do(promeArgs)
		return
	}

	// Bucket level ops
	if BucketName != "" && ObjectName == "" {
		resMp, pkgErr := h.S3pkg.HeadBucket(BucketName, urlParmas)
		promeArgs.Api = "S3pkg.HeadBucket"
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
		Logger.Info("HandleHttpHead()",
			zap.Any("w.Header()", w.Header()))
		go prome.Prometheus.Do(promeArgs)
		return
	}

	errMsg := "Request not supported yet for HEAD."
	HandleErrorRequest(w, h, errMsg)
	return
}
