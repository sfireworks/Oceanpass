package s3_pkg

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type RequestBodyGrants struct {
	Grant []struct {
		Grantee struct {
			DisplayName  string `xml:DisplayName`
			EmailAddress string `xml:EmailAddress`
			ID           string `xml:ID`
			Text         string `xml:",chardata"`
			Xsi          string `xml:"xsi,attr"`
			Type         string `xml:"type,attr"`
			URI          string `xml:URI`
		} `xml:"Grantee"`
		Permission string `xml:Permission`
	} `xml:Grant`
}

type RequestBodyLoggingEnabled struct {
	TargetBucket string            `xml:TargetBucket`
	TargetPrefix string            `xml:TargetPrefix`
	TargetGrants RequestBodyGrants `xml:TargetGrants`
}
type RequestBodyPBL struct {
	LoggingEnabled RequestBodyLoggingEnabled `xml:LoggingEnabled`
}

func (s3pkg *S3Pkg) PutBucketLogging(bucketName string, r *http.Request) (map[string]string, error) {

	var requestPBL RequestBodyPBL
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	err := xml.Unmarshal([]byte(bodyBytes), &requestPBL)

	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.PutBucketLogging", zap.Any("error", err))
		return nil, err
	}
	tmp := requestPBL.LoggingEnabled.TargetGrants.Grant
	targetgrands := make([]types.TargetGrant, len(tmp))
	for i := range tmp {
		var grantee types.Grantee
		grantee.ID = aws.String(tmp[i].Grantee.ID)
		grantee.EmailAddress = aws.String(tmp[i].Grantee.EmailAddress)
		grantee.DisplayName = aws.String(tmp[i].Grantee.DisplayName)
		targetgrands[i].Permission = types.BucketLogsPermission(tmp[i].Permission)
		grantee.Type = types.Type(tmp[i].Grantee.Type)
		targetgrands[i].Grantee = &grantee

	}
	s3PutBucketLoggingInput := &s3.PutBucketLoggingInput{
		Bucket: aws.String(requestPBL.LoggingEnabled.TargetBucket),
		BucketLoggingStatus: &types.BucketLoggingStatus{
			LoggingEnabled: &types.LoggingEnabled{
				TargetBucket: aws.String(requestPBL.LoggingEnabled.TargetBucket),
				TargetPrefix: aws.String(requestPBL.LoggingEnabled.TargetPrefix),
				TargetGrants: targetgrands,
			},
		},
	}
	Logger.Info("S3Pkg.PutBucketLogging()",
		zap.String("endpoint", s3pkg.Endpoint),
		zap.Any("Credentials", s3pkg.AwsConfig.Credentials),
		zap.Any("s3LoV2Input", s3PutBucketLoggingInput))
	result, err := s3pkg.S3Client.PutBucketLogging(context.TODO(), s3PutBucketLoggingInput)
	if err != nil {
		Logger.Error("S3Pkg.PutBucketLogging", zap.Any("error", err))
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("S3Pkg.PutBucketLogging(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
