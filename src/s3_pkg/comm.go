// ///////////////////////////////////////
// 2023 SHAILab Storage all rights reserved
// ///////////////////////////////////////
package s3_pkg

import (
	"context"
	"encoding/xml"
	"net/url"

	// . "oceanpass/src/zaplog"
	zaplog "oceanpass/src/zaplog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/zap"
)

type S3Pkg struct {
	BucketName      string
	S3Client        *s3.Client
	AwsConfig       aws.Config
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	Region          string
	ServiceName     string
	StsToken        string
}

type EnumerationResult struct {
	// CommonPrefixe         []types.CommonPrefix
	CommonPrefixes        []types.CommonPrefix
	Contents              []types.Object
	ContinuationToken     *string
	Delimiter             *string
	EncodingType          types.EncodingType
	IsTruncated           bool
	KeyCount              int32
	MaxKeys               int32
	Name                  *string
	NextContinuationToken *string
	Marker                *string
	Prefix                *string
	StartAfter            *string
	NextMarker            *string
	ResultMetadata        middleware.Metadata
	// noSmithyDocumentSerde
}
type RequestAccessControlPolicy struct {
	XMLName xml.Name `xml:"AccessControlPolicy"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	Owner   struct {
		Text        string `xml:",chardata"`
		DisplayName string `xml:"DisplayName"`
		ID          string `xml:"ID"`
	} `xml:"Owner"`
	AccessControlList struct {
		Text  string `xml:",chardata"`
		Grant []struct {
			Text    string `xml:",chardata"`
			Grantee struct {
				Text         string `xml:",chardata"`
				Xsi          string `xml:"xsi,attr"`
				Type         string `xml:"type,attr"`
				URI          string `xml:"URI"`
				DisplayName  string `xml:"DisplayName"`
				ID           string `xml:"ID"`
				EmailAddress string `xml:"EmailAddress"`
			} `xml:"Grantee"`
			Permission string `xml:"Permission"`
		} `xml:"Grant"`
	} `xml:"AccessControlList"`
}

type AccessControlList struct {
	Grant []types.Grant
}

var Logger = zaplog.Logger

func (s *S3Pkg) New() {

}

func HandleError(err error) {
	// fmt.Println("Error:", err)
	// os.Exit(-1)
}

func GetQueryString(vars url.Values, key string) string {
	if len(vars[key]) > 0 {
		return vars[key][0]
	}
	return ""
}

func (pkg *S3Pkg) GetS3Client() (err error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           pkg.Endpoint,
			SigningRegion: pkg.Region,
		}, nil
	})

	var sdkconfig aws.Config
	sdkconfig, err = config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(pkg.AccessKeyId,
				pkg.AccessKeySecret, pkg.StsToken)))
	if err != nil {
		Logger.Info("GetS3Client()", zap.Any("S3pkg.GetS3Client()", err))
		return err
	}
	pkg.S3Client = s3.NewFromConfig(sdkconfig)
	pkg.AwsConfig = sdkconfig
	return nil
}
