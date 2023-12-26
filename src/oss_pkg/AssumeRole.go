// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Shiqian yan
// ///////////////////////////////////////
package oss_pkg

import (
	"bytes"
	"encoding/xml"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	sts "github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"go.uber.org/zap"
)

func UTCToLocalTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		Logger.Error("UTCToLocalTime()", zap.Any("err", err))
	}
	res := t.Add(8 * time.Hour).Format("2006-01-02 15:04:05")
	return res
}

func (pkg *OssPkg) AssumeRole(roleArn, sessionName, durationSeconds string) ([]byte, int, error) {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(pkg.AccessKeyId, pkg.AccessKeySecret)
	client, err := sts.NewClientWithOptions("cn-shanghai", config, credential)
	if err != nil {
		Logger.Error("AssumeRole().", zap.Any("err", err))
		return nil, -1, err
	}
	request := sts.CreateAssumeRoleRequest()
	request.RoleArn = roleArn
	request.Scheme = "https"
	request.RoleSessionName = sessionName
	// min = 900s  max = 3600s
	if durationSeconds == "" {
		request.DurationSeconds = requests.Integer(pkg.CfgStsDefaultDurationSecs)
	} else {
		request.DurationSeconds = requests.Integer(durationSeconds)
	}
	response, err := client.AssumeRole(request)
	if err != nil {
		Logger.Error("AssumeRole().", zap.Any("err", err))
		return nil, response.GetHttpStatus(), err
	}
	cred := &response.Credentials
	cred.Expiration = UTCToLocalTime(cred.Expiration)
	resXML, err := xml.MarshalIndent(response, " ", " ")
	if err != nil {
		Logger.Error("OSSPkg.AssumeRole(): marshal xml err ",
			zap.Any("error", err))
		return nil, -1, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	return res, -1, nil
}
