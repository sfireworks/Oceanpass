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

func (s3pkg *S3Pkg) PutObjectAcl(bucketName, objectName string,
	r *http.Request) (map[string]string, error) {
	xaa := r.Header.Get("X-Amz-Acl") + r.Header.Get("X-Oss-Object-Acl")
	xagfc := r.Header.Get("X-Amz-Grant-Full-Control")
	xagr := r.Header.Get("X-Amz-Grant-Read")
	cm5 := r.Header.Get("Content-MD5")
	xagwa := r.Header.Get("X-Amz-Grant-Write-Acp")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Logger.Error("S3Pkg.PutObjectAcl(): read body err ",
			zap.Any("error", err))
	}
	var requestPOA RequestAccessControlPolicy

	putObjectAclBody := types.AccessControlPolicy{}

	if len(bodyBytes) != 0 {
		err = xml.Unmarshal([]byte(bodyBytes), &requestPOA)
		if err != nil {
			Logger.Error("S3Pkg.PutObjectAcl(): xml unmarshal err ",
				zap.Any("error", err))
			return nil, err
		}
	}

	tmpowner := types.Owner{}
	tmpowner.ID = aws.String(requestPOA.Owner.ID)
	tmpowner.DisplayName = aws.String(requestPOA.Owner.DisplayName)
	putObjectAclBody.Owner = &tmpowner

	for i := range requestPOA.AccessControlList.Grant {
		tmp := types.Grant{}
		tmpgrantee := types.Grantee{}
		tmp.Permission = types.Permission(requestPOA.AccessControlList.Grant[i].Permission)
		tmpgrantee.Type = types.Type(requestPOA.AccessControlList.Grant[i].Grantee.Type) //aws.String(requestPOA.AccessControlList.Grant[i].Grantee.Type)
		tmpgrantee.ID = aws.String(requestPOA.AccessControlList.Grant[i].Grantee.ID)
		tmpgrantee.URI = aws.String(requestPOA.AccessControlList.Grant[i].Grantee.URI)
		tmpgrantee.DisplayName = aws.String(requestPOA.AccessControlList.Grant[i].Grantee.DisplayName)
		tmpgrantee.EmailAddress = aws.String(requestPOA.AccessControlList.Grant[i].Grantee.EmailAddress)
		tmp.Grantee = &tmpgrantee
		putObjectAclBody.Grants = append(putObjectAclBody.Grants, tmp)
	}

	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Error("S3Pkg.PutObjectAcl:GetS3Client()", zap.Any("error", err))
		return nil, err
	}
	s3PutObjectAclInput := &s3.PutObjectAclInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		//	AccessControlPolicy: &putObjectAclBody,
		ACL:              types.ObjectCannedACL(xaa),
		ContentMD5:       aws.String(cm5),
		GrantFullControl: aws.String(xagfc),
		GrantRead:        aws.String(xagr),
		GrantReadACP:     aws.String(xagwa),
	}
	if len(bodyBytes) != 0 {
		s3PutObjectAclInput.AccessControlPolicy = &putObjectAclBody
	}
	result, err := s3pkg.S3Client.PutObjectAcl(context.TODO(), s3PutObjectAclInput)
	if err != nil {
		Logger.Error("S3Pkg.PutObjectAcl ",
			zap.Any("s3PutObjectAclInput", s3PutObjectAclInput),
			zap.Any("error", err))
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("S3Pkg.PutObjectAcl(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
