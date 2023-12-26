package server

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"oceanpass/src/common"
	"oceanpass/src/config"
	"oceanpass/src/dbops"
	"oceanpass/src/encryption"
	"oceanpass/src/http_handler"
	"oceanpass/src/oss_pkg"
	"oceanpass/src/zaplog"
	"os/user"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TODO: write config to ~/.oceanpass
// FINISHED: ~/.oceanpass/credentials
// TODO: read other configuration from ~/.oceanpass/config
type OcnServer struct {
	// handler    http_handler.HttpHandler
	Config         config.Config
	DBConfig       config.DBConfig
	DirConfig      string
	DirDBConfig    string
	ConfigVariable common.ConfigVariable
	OssDb          *sql.DB
}

var Logger = zaplog.Logger

func (server *OcnServer) Init() {

	server.Config.LoadXMLConfig(server.DirConfig)
	server.ConfigVariable.CfgServiceName = server.Config.ServiceName
	server.ConfigVariable.CfgMode = server.Config.ServerMode
	server.ConfigVariable.CfgStsDefaultDurationSecs = server.Config.StsDefaultDurationSecs

	server.loadConfigEndpoint()
	server.loadConfigCredentials()

	server.DBConfig.LoadXMLDBConfig(server.DirDBConfig)
	server.OssDb = server.DBConfig.ConnDB()
	server.ConfigVariable.OssDb = server.OssDb
	server.ConfigVariable.DBConfig = server.DBConfig

}

func (server *OcnServer) loadConfigEndpoint() {

	Endpoints := server.Config.EndpointConfig.Endpoints
	EndpointsMap := make(map[string]string, len(Endpoints))
	for i := 0; i < len(Endpoints); i++ {
		EndpointsMap[Endpoints[i].ConfigFlag] = Endpoints[i].EndpointUrl
	}
	//Endpoint
	server.ConfigVariable.CfgEndpoint =
		EndpointsMap[server.Config.EndpointConfig.CloudProviderForEndpoint]
	if server.ConfigVariable.CfgEndpoint == "" {
		errMsg := fmt.Sprintf("Endpoint is empty! :%+v", server.Config.EndpointConfig)
		log.Fatalf(errMsg)
	}
	//CloudProviderForEndpoint
	if is, _ := regexp.MatchString(common.K_CLOUD_PROVIDER_OSS,
		server.Config.EndpointConfig.CloudProviderForEndpoint); is {
		server.ConfigVariable.CfgCloudProvider = common.K_CLOUD_PROVIDER_OSS
	} else if is, _ := regexp.MatchString(common.K_CLOUD_PROVIDER_COS,
		server.Config.EndpointConfig.CloudProviderForEndpoint); is {
		server.ConfigVariable.CfgCloudProvider = common.K_CLOUD_PROVIDER_COS
	} else if is, _ := regexp.MatchString(common.K_CLOUD_PROVIDER_OBS,
		server.Config.EndpointConfig.CloudProviderForEndpoint); is {
		server.ConfigVariable.CfgCloudProvider = common.K_CLOUD_PROVIDER_OBS
	} else {
		server.ConfigVariable.CfgCloudProvider = common.K_CLOUD_PROVIDER_OSS
	}

	log.Println()
	log.Printf("AllEndpointConfig:%+v", server.Config.EndpointConfig)
	log.Println("Endpoint:", server.ConfigVariable.CfgEndpoint)
}

func (server *OcnServer) loadConfigCredentials() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	homeDir := currentUser.HomeDir

	oceanpassCfgMapCredentials := config.GetConfigFromFile(homeDir + "/.oceanpass/credentials")
	log.Println("oceanpassCredentials:", oceanpassCfgMapCredentials)

	server.ConfigVariable.CfgAccessKeyId = oceanpassCfgMapCredentials["aws_access_key_id"]
	server.ConfigVariable.CfgAccessKeySecret = oceanpassCfgMapCredentials["aws_secret_access_key"]

}

func (server *OcnServer) RegisterHttpHandler() {
	// All the handler func map for request from client
	var requestHandlers = map[string]func(http.ResponseWriter, *http.Request){
		"/": server.HTTPHandler,
	}
	for map_key, _ := range requestHandlers {
		http.HandleFunc(map_key, requestHandlers[map_key])
	}
}

func (server *OcnServer) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("\t\tINFO    HTTPHandler()\t http.Request:\t %+v ", *r)
	handler := new(http_handler.HttpHandler)
	err := server.InitHttpHandler(handler, &server.ConfigVariable, w, r)
	if err != nil {
		return
	}
	err = server.CheckSignatureIfNecessary(handler, w, r)
	if err != nil {
		return
	}

	method := strings.ToUpper(r.Method)
	// judge which MethodGet
	if method == http.MethodGet {
		w.Header().Set("Content-Type", "application/xml")
		handler.HandleHttpGet(w, r)
	} else if method == http.MethodPost {
		w.Header().Set("Content-Type", "application/xml")
		handler.HandlePost(w, r)
	} else if method == http.MethodHead {
		handler.HandleHttpHead(w, r)
	} else if method == http.MethodPut {
		w.Header().Set("Content-Type", "application/xml")
		handler.HandlePut(w, r)

	} else if method == http.MethodDelete {
		handler.HandleHttpDelete(w, r)

	}
	log.Println("\t\tINFO    HTTPHandler()\t Done. \n ")
}

func (server *OcnServer) CheckSignatureIfNecessary(handler *http_handler.HttpHandler, w http.ResponseWriter, r *http.Request) error {
	if !server.Config.SignatureDisabled {
		userAgent := r.Header.Get("User-Agent")
		var calculateSignature string
		var signature string
		if r.Header.Get("X-Acs-Action") == "AssumeRole" {
			signature = r.URL.Query().Get("Signature")
			calculateSignature = encryption.SignRpcRequest(handler.S3pkg.AccessKeySecret, r)
		} else if r.URL.Query().Get("Signature") != "" {
			signature = r.URL.Query().Get("Signature")
			Expires := r.URL.Query().Get("Expires")
			calculateSignature = encryption.CalculateOSSSignature(r, handler.S3pkg.AccessKeySecret, "OSS", true, Expires)
		} else if r.URL.Query().Get("x-oss-signature") != "" {
			signature = r.URL.Query().Get("x-oss-signature")
			version := r.URL.Query().Get("x-oss-signature-version")
			expires := r.URL.Query().Get("x-oss-expires")
			if version == "OSS" {
				calculateSignature = encryption.CalculateOSSSignature(r, handler.S3pkg.AccessKeySecret, version, true, expires)
			} else if version == "OSS2" {
				calculateSignature = encryption.CalculateOSSSignature(r, handler.S3pkg.AccessKeySecret, version, true, expires)
			} else if version == "OSS4" {
				calculateSignature = encryption.CalculateOSSV4Signature(r, handler.S3pkg.AccessKeySecret, version)
			} else {
				w.WriteHeader(http.StatusForbidden)
				errStruct := common.Error{}
				errStruct.Code = "SignatureMethodDoesNotMatch"
				errStruct.Message = "SignatureMethodDoesNotMatch."
				resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
				byteErr := []byte(resXML)
				w.Write(byteErr)
				err := fmt.Sprintf("error: %+v", errStruct)
				Logger.Error("CheckSignatureIfNecessary()：SignatureMethodDoesNotMatch",
					zap.String("err", err), zap.String("signature", signature), zap.String("calculateSignature", calculateSignature))
				return fmt.Errorf(err)
			}
		} else if r.URL.Query().Get("X-Amz-Signature") != "" {
			signature = r.URL.Query().Get("X-Amz-Signature")
			calculateSignature = encryption.CalculateAwsPreSignature(handler.S3pkg.AccessKeySecret, r, handler.S3pkg.ServiceName)
		} else if strings.Contains(userAgent, "aliyun") || strings.Contains(userAgent, "Alibaba") {
			version := http_handler.GetOssSignatureVersion(r)
			if version == "OSS" {
				calculateSignature = encryption.CalculateOSSSignature(r, handler.S3pkg.AccessKeySecret, version, false, "")
				signature = http_handler.GetOssSignatureV1(r)
			} else if version == "OSS2" {
				calculateSignature = encryption.CalculateOSSSignature(r, handler.S3pkg.AccessKeySecret, version, false, "")
				signature = http_handler.GetOssSignatureV2(r)
			} else if version == "OSS4" {
				calculateSignature = encryption.CalculateOSSV4Signature(r, handler.S3pkg.AccessKeySecret, version)
				signature = http_handler.GetOssSignatureV4(r)
			} else {
				w.WriteHeader(http.StatusForbidden)
				errStruct := common.Error{}
				errStruct.Code = "SignatureMethodDoesNotMatch"
				errStruct.Message = "SignatureMethodDoesNotMatch."
				resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
				byteErr := []byte(resXML)
				w.Write(byteErr)
				err := fmt.Sprintf("error: %+v", errStruct)
				Logger.Error("CheckSignatureIfNecessary()：SignatureMethodDoesNotMatch",
					zap.String("err", err), zap.String("signature", signature), zap.String("calculateSignature", calculateSignature))
				return fmt.Errorf(err)
			}
		} else if strings.Contains(userAgent, "aws") ||
			strings.Contains(userAgent, "Boto3") {
			region := http_handler.GetRegion(r)
			serviceNameFromHttp := http_handler.GetServiceName(r)
			handler.S3pkg.Region = region
			handler.S3pkg.ServiceName = serviceNameFromHttp
			handler.Osspkg.ServiceName = serviceNameFromHttp
			Logger.Info("HTTPHandler()",
				zap.String("region", region),
				zap.String("ServiceName", serviceNameFromHttp))

			calculateSignature = encryption.CalculateS3V4Signature(handler.S3pkg.AccessKeySecret, r, handler.S3pkg.Region, handler.S3pkg.ServiceName)
			signature = http_handler.GetS3SignatureV4(r)
		} else {
			w.WriteHeader(http.StatusForbidden)
			errStruct := common.Error{}
			errStruct.Code = "SignatureMethodDoesNotMatch"
			errStruct.Message = "SignatureMethodDoesNotMatch."
			resXML, _ := xml.MarshalIndent(errStruct, " ", " ")
			byteErr := []byte(resXML)
			w.Write(byteErr)
			err := fmt.Sprintf("error: %+v", errStruct)
			Logger.Error("CheckSignatureIfNecessary()：SignatureMethodDoesNotMatch",
				zap.String("err", err), zap.String("signature", signature), zap.String("calculateSignature", calculateSignature))
			return fmt.Errorf(err)
		}
		if signature != calculateSignature {
			// if the program crash on here, tell shiqianyan to fix it.
			BucketName := http_handler.GetBucketName(r)
			ocnOssError := oss_pkg.Error{}
			ocnOssError.HttpsResponseErrorStatusCode = fmt.Sprintf("%d", http.StatusForbidden)
			ocnOssError.Code = "SignatureDoesNotMatch"
			ocnOssError.Message = "The request signature we calculated does not match the signature you provided. Check your key and signing method."
			re, _ := regexp.Compile("([^://]+)$")
			endpoint := re.FindString(handler.Osspkg.Endpoint)
			ocnOssError.HostId = BucketName + "." + endpoint
			http_handler.WriteErrorToHeader(ocnOssError, w)
			http_handler.WriteHeaderStatusCode(ocnOssError, w)
			http_handler.WriteErrorToBody(ocnOssError, w)
			err := fmt.Sprintf("error: %+v", ocnOssError)
			Logger.Error("CheckSignatureIfNecessary()",
				zap.String("err", err), zap.String("signature", signature), zap.String("calculateSignature", calculateSignature))
			return fmt.Errorf(err)

		}

	}
	return nil
}

func (server *OcnServer) InitHttpHandler(handler *http_handler.HttpHandler, comCfgGloVar *common.ConfigVariable, w http.ResponseWriter, r *http.Request) error {
	handler.S3pkg.AccessKeyId = comCfgGloVar.CfgAccessKeyId
	handler.S3pkg.AccessKeySecret = comCfgGloVar.CfgAccessKeySecret
	handler.S3pkg.Endpoint = comCfgGloVar.CfgEndpoint
	handler.Osspkg.AccessKeyId = comCfgGloVar.CfgAccessKeyId
	handler.Osspkg.AccessKeySecret = comCfgGloVar.CfgAccessKeySecret
	handler.Osspkg.Endpoint = comCfgGloVar.CfgEndpoint
	handler.Osspkg.CfgStsDefaultDurationSecs = comCfgGloVar.CfgStsDefaultDurationSecs
	stsToken := http_handler.GetStsToken(r)
	handler.S3pkg.StsToken = stsToken
	handler.Osspkg.StsToken = stsToken
	handler.DBConfig = comCfgGloVar.DBConfig
	handler.OssDb = comCfgGloVar.OssDb
	handler.CloudProvider = comCfgGloVar.CfgCloudProvider

	handler.ClientProvider = common.K_CLIENT_PROVIDER_AMAZON
	if is, _ :=
		regexp.MatchString(common.K_CLIENT_PROVIDER_ALIYUN,
			strings.ToLower(r.Header.Get("User-Agent"))); is {
		handler.ClientProvider = common.K_CLIENT_PROVIDER_ALIYUN
	}

	Logger.Info("HTTPHandler()",
		zap.String("Config mode", comCfgGloVar.CfgMode))
	userAgent := r.Header.Get("User-Agent")
	var akIdFromHttp string
	if r.Header.Get("X-Acs-Action") == "AssumeRole" {
		//sts case
		akIdFromHttp = r.URL.Query().Get("AccessKeyId")
	} else if r.URL.Query().Get("OSSAccessKeyId") != "" {
		akIdFromHttp = r.URL.Query().Get("OSSAccessKeyId")
	} else if r.URL.Query().Get("x-oss-access-key-id") != "" {
		akIdFromHttp = r.URL.Query().Get("x-oss-access-key-id")
	} else if r.URL.Query().Get("X-Amz-Credential") != "" {
		akIdFromHttp = http_handler.GetAccessKeyIdFromXAmzCredential(r)
	} else if strings.Contains(userAgent, "aliyun") || strings.Contains(userAgent, "Alibaba") {
		version := http_handler.GetOssSignatureVersion(r)
		if version == "OSS" {
			akIdFromHttp = http_handler.GetAccessKeyIdV1(r)
		} else if version == "OSS2" {
			akIdFromHttp = http_handler.GetAccessKeyIdV2(r)
		} else if version == "OSS4" {
			akIdFromHttp = http_handler.GetAccessKeyIdV4(r)
		}
	} else if strings.Contains(userAgent, "aws") ||
		strings.Contains(userAgent, "Boto3") {
		akIdFromHttp = http_handler.GetAccessKeyIdV4(r)
	}
	if akIdFromHttp != comCfgGloVar.CfgAccessKeyId && comCfgGloVar.CfgMode == "xml" {
		BucketName := http_handler.GetBucketName(r)
		ocnOssError := oss_pkg.Error{}
		ocnOssError.HttpsResponseErrorStatusCode = fmt.Sprintf("%d", http.StatusForbidden)
		ocnOssError.Code = "InvalidAccessKeyId"
		ocnOssError.Message = "The OSS Access Key Id you provided does not exist in our records."
		re, _ := regexp.Compile("([^://]+)$")
		endpoint := re.FindString(handler.Osspkg.Endpoint)
		ocnOssError.HostId = BucketName + "." + endpoint
		http_handler.WriteErrorToHeader(ocnOssError, w)
		http_handler.WriteHeaderStatusCode(ocnOssError, w)
		http_handler.WriteErrorToBody(ocnOssError, w)
		err := fmt.Sprintf("error: %+v", ocnOssError)
		Logger.Error("InitHttpHandler()",
			zap.String("err", err), zap.String("CfgAccessKeyId", comCfgGloVar.CfgAccessKeyId), zap.String("akIdFromHttp", akIdFromHttp))
		return fmt.Errorf(err)
	}
	if comCfgGloVar.CfgMode == "db" {
		akSecretFromDB, err := dbops.GetSecretByIdFromDB(akIdFromHttp, server.OssDb, server.DBConfig.DbBase.TableName)
		if err != nil {
			BucketName := http_handler.GetBucketName(r)
			ocnOssError := oss_pkg.Error{}
			ocnOssError.HttpsResponseErrorStatusCode = fmt.Sprintf("%d", http.StatusForbidden)
			ocnOssError.Code = "InvalidAccessKeyId"
			ocnOssError.Message = "The OSS Access Key Id you provided does not exist in our records."
			re, _ := regexp.Compile("([^://]+)$")
			endpoint := re.FindString(handler.Osspkg.Endpoint)
			ocnOssError.HostId = BucketName + "." + endpoint
			http_handler.WriteErrorToHeader(ocnOssError, w)
			http_handler.WriteHeaderStatusCode(ocnOssError, w)
			http_handler.WriteErrorToBody(ocnOssError, w)
			Logger.Error("dbops.GetSecretByIdFromDB", zap.Any("err", err))
			return err
		}
		Logger.Info("HTTPHandler()", zap.String("akIdFromHttp", akIdFromHttp), zap.String("akSecretFromDB", akSecretFromDB))
		handler.S3pkg.AccessKeyId = akIdFromHttp
		handler.S3pkg.AccessKeySecret = akSecretFromDB
		handler.Osspkg.AccessKeyId = akIdFromHttp
		handler.Osspkg.AccessKeySecret = akSecretFromDB
	}
	return nil
}

func (server *OcnServer) DeleteExpireItemInDB() {
	for {
		time.Sleep(1 * time.Minute)
		err := dbops.DeleteExpireItemInDB(server.OssDb, server.DBConfig.DbBase.TableName)
		if err != nil {
			Logger.Error("dbops.DeleteExpireItemInDB", zap.Any("err", err))
		}
	}
}
