// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package s3_pkg

import (
	"context"
	"fmt"
	"oceanpass/src/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type GetBucketAclOutput struct {
	AccessControlList AccessControlList
	Owner             *types.Owner
}

func (pkg *S3Pkg) GetBucketAcl(bucketName string,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("S3Pkg.GetBucketAcl(): GetOssClient()", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetBucketAclInput := &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	}

	s3GoaRes, err := pkg.S3Client.GetBucketAcl(context.TODO(), s3GetBucketAclInput)
	if err != nil {
		Logger.Error("S3Pkg.GetBucketAcl()",
			zap.String("bucketName", bucketName),
			zap.Any("error", err))
		return nil, nil, err
	}
	respGetRawResponse := mid.GetRawResponse(s3GoaRes.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	accessControlList := AccessControlList{Grant: s3GoaRes.Grants}
	goaRes := GetBucketAclOutput{
		AccessControlList: accessControlList,
		Owner:             s3GoaRes.Owner,
	}

	res, err := common.MarshalIndent(goaRes)
	if err != nil {
		Logger.Error("S3Pkg.GetBucketAcl(): marshal xml err ",
			zap.Any("error", err))
	}

	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	Logger.Info("S3Pkg.GetBucketAcl(): ok.",
		zap.Any("result", string(res)))

	return res, resMp, nil
}
