package s3_pkg

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awshttp "github.com/aws/smithy-go/transport/http"

	"go.uber.org/zap"
)

func (s3pkg *S3Pkg) PutBucketAcl(bucketName string, r *http.Request) (map[string]string, error) {
	log.Println(
		"[S3Pkg] Call PutBucketAcl.",
		" endpoint:", s3pkg.Endpoint,
		"\n bucketName:", bucketName,
		"\n ")

	bca := r.Header.Get("X-Amz-Acl") + r.Header.Get("X-Oss-Acl")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read body err")
	}
	//log.Println("bca,body. == ",bca,bodyBytes)
	requestPBA := RequestAccessControlPolicy{}
	if len(bodyBytes) != 0 {
		err = xml.Unmarshal([]byte(bodyBytes), &requestPBA)
		if err != nil {
			log.Println("xml unmarshal err,", err)
			return nil, err
		}

	}
	log.Println("requestPBA == ", requestPBA)
	putBucketAclBody := types.AccessControlPolicy{}
	tmpowner := types.Owner{}
	tmpowner.ID = aws.String(requestPBA.Owner.ID)
	tmpowner.DisplayName = aws.String(requestPBA.Owner.DisplayName)
	putBucketAclBody.Owner = &tmpowner

	for i := range requestPBA.AccessControlList.Grant {
		tmp := types.Grant{}
		tmpgrantee := types.Grantee{}
		tmp.Permission = types.Permission(requestPBA.AccessControlList.Grant[i].Permission)
		tmpgrantee.Type = types.Type(requestPBA.AccessControlList.Grant[i].Grantee.Type) //aws.String(requestPOA.AccessControlList.Grant[i].Grantee.Type)
		tmpgrantee.ID = aws.String(requestPBA.AccessControlList.Grant[i].Grantee.ID)
		tmpgrantee.URI = aws.String(requestPBA.AccessControlList.Grant[i].Grantee.URI)
		tmpgrantee.DisplayName = aws.String(requestPBA.AccessControlList.Grant[i].Grantee.DisplayName)
		tmpgrantee.EmailAddress = aws.String(requestPBA.AccessControlList.Grant[i].Grantee.EmailAddress)
		tmp.Grantee = &tmpgrantee
		putBucketAclBody.Grants = append(putBucketAclBody.Grants, tmp)
	}
	log.Println("putBucketAclBody == ", putBucketAclBody)
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("PutBucketAcl", zap.Any("error", err))
		return nil, err
	}

	s3PutBucketAclInput := &s3.PutBucketAclInput{
		Bucket: aws.String(bucketName),
		//		ACL:    types.BucketCannedACL(bca),
	}
	if len(bca) != 0 {
		s3PutBucketAclInput.ACL = types.BucketCannedACL(bca)
	}
	if len(bodyBytes) != 0 {
		s3PutBucketAclInput.AccessControlPolicy = &putBucketAclBody
	}
	result, err := s3pkg.S3Client.PutBucketAcl(context.TODO(), s3PutBucketAclInput)
	if err != nil {
		log.Println("[ERROR PutBucketAcl]: ", err)
		return nil, err
	}

	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("<s3>.PutBucketAcl(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
