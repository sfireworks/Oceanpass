// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package common

import (
	"database/sql"
	"oceanpass/src/config"
)

type ConfigVariable struct {
	CfgAccessKeyId            string
	CfgAccessKeySecret        string
	CfgEndpoint               string
	CfgMode                   string
	CfgServiceName            string
	CfgCloudProvider          string
	CfgStsDefaultDurationSecs string
	OssDb                     *sql.DB
	DBConfig                  config.DBConfig
}

type Authorization struct {
	SignMethod  string
	AccessKeyId string
	Signature   string

	//awscli
	AwsCredential      string
	AwsAuthDate        string
	AwsAuthRegion      string
	AwsAuthServiceName string
	AwsAuthRequestName string
	AwsSignedHeaders   string

	//oss
	OssAccessKeyIdAndSignature string
}

var ErrorCodeMessageMap = map[string]string{
	"NoSuchKey": "The specified key does not exist.",
	"BucketAlreadyExists": "The requested bucket name is not available. " +
		"The bucket namespace is shared by all users of the system. " +
		"Please select a different name and try again.",
	"NoSuchBucket": "",
}
