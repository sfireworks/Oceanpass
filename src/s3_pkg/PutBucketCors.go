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

type RequestBodyCORSRule struct {
	AllowedOrigin []string `xml:"AllowedOrigin"`
	AllowedHeader []string `xml:"AllowedHeader"`
	AllowedMethod []string `xml:"AllowedMethod"`
	ExposeHeader  []string `xml:"ExposeHeader"`
	ID            string   `xml:"ID"`
	MaxAgeSeconds int      `xml:"MaxAgeSeconds"`
}
type RequestBodyForPutBucketCors struct {
	Name     string                `xml:"CORSConfiguration"`
	CORSRule []RequestBodyCORSRule `xml:"CORSRule"`
}

var requestPBC RequestBodyForPutBucketCors

func (s3pkg *S3Pkg) PutBucketCors(bucketName string, r *http.Request) (map[string]string, error) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	err := xml.Unmarshal([]byte(bodyBytes), &requestPBC)
	if err != nil {
		log.Println("xml unmarshal err,", err)
		return nil, err
	}

	var CorsRuleList []types.CORSRule
	for i := range requestPBC.CORSRule {
		tmp := types.CORSRule{}
		tmp.AllowedOrigins = requestPBC.CORSRule[i].AllowedOrigin
		tmp.AllowedHeaders = requestPBC.CORSRule[i].AllowedHeader
		tmp.AllowedMethods = requestPBC.CORSRule[i].AllowedMethod
		tmp.ID = aws.String(requestPBC.CORSRule[i].ID)
		tmp.ExposeHeaders = requestPBC.CORSRule[i].ExposeHeader
		tmp.MaxAgeSeconds = int32(requestPBC.CORSRule[i].MaxAgeSeconds)
		CorsRuleList = append(CorsRuleList, tmp)
	}
	for i := range CorsRuleList {
		log.Println(CorsRuleList[i])
	}
	requestPBC = RequestBodyForPutBucketCors{}
	err = s3pkg.GetS3Client()
	if err != nil {
		Logger.Info("PutBucketCors", zap.Any("error", err))
		return nil, err
	}

	s3PutBucketCorsInput := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucketName),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: CorsRuleList,
		},
	}
	CorsRuleList = []types.CORSRule{}
	result, err := s3pkg.S3Client.PutBucketCors(context.TODO(), s3PutBucketCorsInput)
	if err != nil {
		log.Println("[ERROR PutbucketCors]: ", err)
		return nil, err
	}
	respGetRawResponse := middleware.GetRawResponse(result.ResultMetadata)
	httpResponse := respGetRawResponse.(*awshttp.Response)
	resMp := make(map[string]string)
	for key, _ := range httpResponse.Header {
		resMp[key] = httpResponse.Header.Get(key)
	}
	Logger.Info("<s3>.PutBucketCors(): ok.", zap.Any("result", resMp))
	return resMp, nil
}
