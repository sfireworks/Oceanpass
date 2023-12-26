// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package http_handler

import (
	"encoding/xml"
	"net/http"
	"oceanpass/src/oss_pkg"
	"oceanpass/src/s3_pkg"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"strconv"

	"go.uber.org/zap"
)

func WriteErrorToBody(ocnErr interface{}, w http.ResponseWriter) (err error) {
	xmlRes, xmlErr := xml.MarshalIndent(ocnErr, " ", " ")
	if xmlErr != nil {
		Logger.Error("WriteError(): marshal xml error: ",
			zap.Any("xmlRes", xmlRes), zap.Any("error", xmlErr))
		errStruct := Error{
			Code:    "MarshalIndent(errStruct)Error",
			Message: xmlErr.Error(),
		}
		resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
		byteErr := []byte(resXML)
		w.Write(byteErr)
		w.WriteHeader(http.StatusInternalServerError) //500
		return xmlErr
	}
	byteErr := []byte(xmlRes)
	w.Write([]byte(xml.Header))
	w.Write(byteErr)
	w.Write([]byte("\n\r"))
	Logger.Info("WriteErrorToBody(): ", zap.Any("byteErr", string(byteErr)))

	return nil
}

func WriteErrorToHeader(ocnErr interface{}, w http.ResponseWriter) {
	resMp := make(map[string]string)
	switch ocnErr.(type) {
	case s3_pkg.Error:
		resMp["x-oss-request-id"] = ocnErr.(s3_pkg.Error).RequestId
		resMp["x-amz-request-id"] = ocnErr.(s3_pkg.Error).RequestID
		resMp["x-oss-host-id"] = ocnErr.(s3_pkg.Error).HostId
		resMp["x-amz-host-id"] = ocnErr.(s3_pkg.Error).HostID
	case oss_pkg.Error:
		resMp["x-oss-request-id"] = ocnErr.(oss_pkg.Error).RequestId
		resMp["x-oss-host-id"] = ocnErr.(oss_pkg.Error).HostId
	default:
		resMp["x-oss-request-id"] = ocnErr.(s3_pkg.Error).RequestId
		resMp["x-amz-request-id"] = ocnErr.(s3_pkg.Error).RequestID
		resMp["x-oss-host-id"] = ocnErr.(s3_pkg.Error).HostId
		resMp["x-amz-host-id"] = ocnErr.(s3_pkg.Error).HostID
	}

	// WriteHeader
	for resKey, resVal := range resMp {
		w.Header().Set(resKey, resVal)
	}
	Logger.Info("WriteErrorToHeader(): ", zap.Any("w.Header()", w.Header()))

}

func WriteHeaderStatusCode(ocnErr interface{}, w http.ResponseWriter) (err error) {
	var httpStatusCode int
	var httpStatusCodeErr error
	switch ocnErr.(type) {
	case s3_pkg.Error:
		httpStatusCode, httpStatusCodeErr =
			strconv.Atoi(ocnErr.(s3_pkg.Error).HttpsResponseErrorStatusCode)
	case oss_pkg.Error:
		httpStatusCode, httpStatusCodeErr =
			strconv.Atoi(ocnErr.(oss_pkg.Error).HttpsResponseErrorStatusCode)
	default:
		httpStatusCode, httpStatusCodeErr =
			strconv.Atoi(ocnErr.(s3_pkg.Error).HttpsResponseErrorStatusCode)
	}
	if httpStatusCodeErr != nil {
		Logger.Error("WriteHeaderStatusCode(): Atoi: ",
			zap.Any("httpStatusCode", httpStatusCode),
			zap.Any("httpCodeErr", httpStatusCodeErr))
		errStruct := Error{
			Code:    "Atoi(httpStatusCode)Error",
			Message: httpStatusCodeErr.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError) //500
		resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
		byteErr := []byte(resXML)
		w.Write(byteErr)
		return httpStatusCodeErr
	}
	w.WriteHeader(httpStatusCode)
	Logger.Info("WriteHeaderStatusCode(): ",
		zap.Any("httpStatusCode", httpStatusCode))
	return nil
}
