// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package s3_pkg

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	mid "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	awshttp "github.com/aws/smithy-go/transport/http"
	"go.uber.org/zap"
)

type GetObjectAclOutput struct {
	AccessControlList AccessControlList
	Owner             *types.Owner
	RequestCharged    types.RequestCharged
	ResultMetadata    middleware.Metadata
	// noSmithyDocumentSerde
}

func (pkg *S3Pkg) GetObjectAcl(
	bucketName, objectName string,
	urlQuery url.Values,
) (bodyBytes []byte, headerMap map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.GetObjectAcl", zap.Any("error", err))
		return nil, nil, err
	}

	s3GetObjectAclInput := &s3.GetObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	if GetQueryString(urlQuery, "versionId") != "" {
		versionId := GetQueryString(urlQuery, "versionId")
		s3GetObjectAclInput.VersionId = &versionId
	}

	s3GoaRes, err := pkg.S3Client.GetObjectAcl(context.TODO(), s3GetObjectAclInput)
	if err != nil {
		Logger.Error("S3Pkg.GetObjectAcl ",
			zap.Any("s3GetObjectAclInput", s3GetObjectAclInput),
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
	goaRes := GetObjectAclOutput{
		AccessControlList: accessControlList,
		Owner:             s3GoaRes.Owner,
		RequestCharged:    s3GoaRes.RequestCharged,
		ResultMetadata:    s3GoaRes.ResultMetadata,
	}

	resXML, err := xml.MarshalIndent(goaRes, " ", " ")
	if err != nil {
		Logger.Error("S3Pkg.GetObjectAcl(): marshal xml err ",
			zap.Any("error", err))
		return nil, nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()
	resMp["Content-Length"] = fmt.Sprintf("%d", len(res))
	Logger.Info("S3Pkg.GetObjectAcl(): ok.", zap.Any("result", res))

	return res, resMp, nil
}
