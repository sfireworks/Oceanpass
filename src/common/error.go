// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin, Shiqian Yan
// ///////////////////////////////////////
package common

import (
	"encoding/xml"
	"net/http"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type Error struct {
	Code                         string
	Message                      string
	OperationErrorS3             string
	HttpsResponseErrorStatusCode string
	RequestID                    string
	HostID                       string
}

func HandleAwsErrorReturn(apiErr error) (ocnErr Error) {
	Logger.Info("HandleAwsErrorReturn()", zap.Any("apiErr", apiErr))

	apiErrSlice := strings.Split(apiErr.Error(), ", ")
	apiErrMap := make(map[string]string, len(apiErrSlice))
	for _, v := range apiErrSlice {
		val := strings.Split(v, ": ")
		apiErrMap[val[0]] = val[1]

		if is, _ := regexp.MatchString("operation", val[0]); is {
			ocnErr.OperationErrorS3 = val[1]
		} else if is, _ := regexp.MatchString("StatusCode", val[0]); is {
			ocnErr.HttpsResponseErrorStatusCode = val[1]
		} else if is, _ := regexp.MatchString("RequestID", val[0]); is {
			ocnErr.RequestID = val[1]
		} else if is, _ := regexp.MatchString("HostID", val[0]); is {
			ocnErr.HostID = val[1]
		} else if is, _ := regexp.MatchString("api error", val[0]); is {
			apiError := strings.Split(val[0], "error ")
			ocnErr.Code = apiError[1]
			ocnErr.Message = val[1]
		}
	}

	return ocnErr
}

func WriteError(ocnErr Error, w http.ResponseWriter) (err error) {
	httpStatusCode, httpStatusCodeErr := strconv.Atoi(ocnErr.HttpsResponseErrorStatusCode)
	if httpStatusCodeErr != nil {
		Logger.Error("WriteError(): Atoi: ",
			zap.Any("httpStatusCode", httpStatusCode),
			zap.Any("httpCodeErr", httpStatusCodeErr))
		w.WriteHeader(http.StatusInternalServerError) //500
		errStruct := Error{
			Code:    "Atoi(httpStatusCode)Error",
			Message: httpStatusCodeErr.Error(),
		}
		resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
		byteErr := []byte(resXML)
		w.Write(byteErr)
		return httpStatusCodeErr
	}

	w.WriteHeader(httpStatusCode)
	Logger.Info("WriteError()", zap.Any("ocnErr", ocnErr))
	errStruct := Error{
		Code:    ocnErr.Code,
		Message: ocnErr.Message,
	}
	xmlRes, xmlErr := xml.MarshalIndent(errStruct, " ", " ")
	if xmlErr != nil {
		Logger.Error("WriteError(): marshal xml error: ",
			zap.Any("xmlRes", xmlRes), zap.Any("error", xmlErr))
		w.WriteHeader(http.StatusInternalServerError) //500
		errStruct := Error{
			Code:    "MarshalIndent(errStruct)Error",
			Message: xmlErr.Error(),
		}
		resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
		byteErr := []byte(resXML)
		w.Write(byteErr)
		return xmlErr
	}
	byteErr := []byte(xmlRes)
	w.Write(byteErr)

	return nil
}

func HandleErrorReturn(apiErr error, w http.ResponseWriter) (err error) {
	ocnErr := HandleAwsErrorReturn(apiErr)
	err = WriteError(ocnErr, w)
	return err
}
