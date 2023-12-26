// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package s3_pkg

import (
	"log"
	"regexp"

	//"oceanpass/src/common"

	"go.uber.org/zap"
)

type Error struct {
	Code                         string
	Message                      string
	OperationErrorS3             string
	HttpsResponseErrorStatusCode string
	// for amazon client (awscli / aws sdk)
	RequestID string
	HostID    string
	// for aliyun client (ossutil64 / oss sdk)
	RequestId string
	HostId    string
}

/*
The awsS3sdk error message string is like:
"operation error S3: PutObjectAcl, https response error StatusCode: 400, RequestID: 642260093D168E3930C86979, HostID: csun-meta-bkt-1.oss-cn-shanghai-internal.aliyuncs.com, api error MissingSecurityHeader: Your request was missing a required header"

"error": "operation error S3: ListObjects, https response error StatusCode: 404, RequestID: 643F62A46814B03534D56B55, HostID: ocntest-1681879570-unormal-list-objects-bucket.oss-cn-shanghai-internal.aliyuncs.com, NoSuchBucket:

This function is to extract the "Code", "Message", "OperationErrorS3", "HttpsResponseErrorStatusCode"
and "RequestID"/"RequestId", "HostID"/"HostId" in this message.
*/
func HandleErrorReturn(apiErr error) (ocnErr Error) {
	sub := "operation.+: (.*?), .+StatusCode: (.*?), " +
		"RequestID: (.*?), HostID: (.*?),.+?(api error (.*?): (.*?))?$"

	re1, err := regexp.Compile(sub)
	if err != nil {
		log.Fatalln("regexp Compile failed!")
	}
	matchArr := re1.FindStringSubmatch(apiErr.Error())
	if len(matchArr) < 5 {
		log.Fatalf("regexp apiErr len mismatch!, %d, %s", len(matchArr), apiErr.Error())
	}
	ocnErr.OperationErrorS3 = matchArr[1]
	ocnErr.HttpsResponseErrorStatusCode = matchArr[2]
	ocnErr.RequestID = matchArr[3]
	ocnErr.RequestId = matchArr[3]
	ocnErr.HostID = matchArr[4]
	ocnErr.HostId = matchArr[4]
	if len(matchArr) >= 8 {
		ocnErr.Code = matchArr[6]
		ocnErr.Message = matchArr[7]
	}
	if ocnErr.Code == "" && ocnErr.Message == "" {
		sub := "operation.+: (.*?), .+StatusCode: (.*?), " +
			"RequestID: (.*?), HostID: (.*?),.+?((.*?): (.*?))?$"

		re1, err := regexp.Compile(sub)
		if err != nil {
			log.Fatalln("regexp Compile failed!")
		}
		matchArr := re1.FindStringSubmatch(apiErr.Error())
		if len(matchArr) < 5 {
			log.Fatalf("regexp apiErr len mismatch!, %d, %s", len(matchArr), apiErr.Error())
		}
		if len(matchArr) >= 8 {
			ocnErr.Code = matchArr[6]
			ocnErr.Message = matchArr[7]
		}
		if ocnErr.Code == "NoSuchKey" {
			ocnErr.Message = "The specified key does not exist."
		}
		if ocnErr.Code == "NoSuchBucket" {
			ocnErr.Message = "The specified bucket does not exist."
		}
	}
	Logger.Info("s3_pkg.HandleErrorReturn()", zap.Any("ocnErr", ocnErr))
	return ocnErr
}
