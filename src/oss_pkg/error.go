// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package oss_pkg

import (
	"fmt"
	"log"
	"regexp"

	"go.uber.org/zap"
)

type Error struct {
	Code                         string
	Message                      string
	HttpsResponseErrorStatusCode string
	// only  for aliyun client (ossutil64 / oss sdk)
	RequestId string
	HostId    string
}

/*
The Oss error message string is like "oss: service returned error: StatusCode=409,
ErrorCode=FileAlreadyExists, ErrorMessage=\"The object you specified already
exists and can not be overwritten.\", RequestId=641D6D20DAC9123233C770C5"

This function is to extract the "StatusCode", "ErrorCode", "ErrorMessage",
and "RequestId" in this message.
*/
func HandleErrorReturn(apiErr error) (ocnErr Error) {
	sub := "StatusCode=(.*?), ErrorCode=(.*?), ErrorMessage=\"(.*?)\", " +
		"RequestId=(.*?)(, HostId=(.*?))?$"
	re, err := regexp.Compile(sub)
	if err != nil {
		log.Fatalln("regexp Compile failed!")
	}
	matchArr := re.FindStringSubmatch(apiErr.Error())

	if len(matchArr) < 5 {
		log.Fatalf("regexp apiErr len %d mismatch!, %s", len(matchArr), apiErr.Error())
	}
	ocnErr.HttpsResponseErrorStatusCode = matchArr[1]
	ocnErr.Code = matchArr[2]
	ocnErr.Message = matchArr[3]
	ocnErr.RequestId = matchArr[4]
	if len(matchArr) > 6 {
		ocnErr.HostId = matchArr[6]
	}
	Logger.Info("oss_pkg.HandleErrorReturn()", zap.Any("ocnErr", ocnErr))
	return ocnErr
}

func HandleAssumeRoleErrorReturn(apiErr error, code int) (ocnErr Error) {
	sub := "ErrorCode:(.*?)\nRecommend:(.*?)\nRequestId:(.*?)\nMessage:(.*?)\n"
	re, err := regexp.Compile(sub)
	if err != nil {
		log.Fatalln("regexp Compile failed!")
	}
	matchArr := re.FindStringSubmatch(apiErr.Error())
	ocnErr.HttpsResponseErrorStatusCode = fmt.Sprintf("%d", code)
	ocnErr.Code = matchArr[1]
	ocnErr.RequestId = matchArr[3]
	ocnErr.Message = matchArr[4]
	Logger.Info("oss_pkg.HandleErrorReturn()", zap.Any("ocnErr", ocnErr))
	return ocnErr
}
