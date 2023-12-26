// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Shiqian Yan
// ///////////////////////////////////////
package encryption

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io/ioutil"
	"net/http"
	. "oceanpass/src/zaplog"
	_ "oceanpass/src/zaplog"
	"sort"
	"strings"

	"go.uber.org/zap"
)

var timeLen = 8

func makeHash(hash hash.Hash, b []byte) []byte {
	hash.Reset()
	hash.Write(b)
	return hash.Sum(nil)
}

func HMACSHA256(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func GetCanonicalString(r *http.Request) (string, string, string) {
	var res strings.Builder
	res.WriteString(r.Method + "\n")
	reqURISlice := strings.Split(r.RequestURI, "?")
	path := reqURISlice[0]
	res.WriteString(path + "\n")
	if len(reqURISlice) > 1 {
		args := strings.Split(reqURISlice[1], "&")
		for i := 0; i < len(args); i++ {
			if !strings.Contains(args[i], "=") {
				args[i] = args[i] + "="
			}
		}
		sort.Strings(args)
		argsString := strings.Join(args, "&")
		res.WriteString(argsString + "\n")
	} else {
		res.WriteString("\n")
	}
	authorizations := strings.Split(r.Header.Get("Authorization"), ",")
	credential := strings.Split(authorizations[0], "/")
	credentialScope := strings.Join(credential[1:], "/")
	signedHeaders := authorizations[len(authorizations)-2][len("signedHeaders=")+1:]
	signedHeadersArr := strings.Split(signedHeaders, ";")
	for i := 0; i < len(signedHeadersArr); i++ {
		if signedHeadersArr[i] == "host" {
			res.WriteString(signedHeadersArr[i] + ":" + r.Host + "\n")
		} else {
			res.WriteString(signedHeadersArr[i] + ":" + r.Header.Get(signedHeadersArr[i]) + "\n")
		}
	}
	res.WriteRune('\n')
	res.WriteString(signedHeaders + "\n")
	payloadHash, err := CalculatePayloadHash(r)
	if err != nil {
		Logger.Error("GetCanonicalString()", zap.Any("err", err))
		return "", "", ""
	}
	hashFromHttp := r.Header.Get("X-Amz-Content-Sha256")
	if hashFromHttp != "" && hashFromHttp != payloadHash {
		Logger.Error("GetCanonicalString()", zap.Any("err", "payloadHash is not equal"))
		return "", "", ""
	}
	res.WriteString(payloadHash)
	return res.String(), r.Header.Get("X-Amz-Date"), credentialScope
}

func GetPreCanonicalString(r *http.Request) (string, string, string) {
	var canonicalRequestString, xTime, credentialScope string
	var res strings.Builder
	res.WriteString(r.Method + "\n")
	reqURISlice := strings.Split(r.RequestURI, "?")
	path := reqURISlice[0]
	res.WriteString(path + "\n")
	if len(reqURISlice) > 1 {
		args := strings.Split(reqURISlice[1], "&")
		for i := 0; i < len(args); i++ {
			if strings.Contains(args[i], "X-Amz-Signature") {
				args = append(args[:i], args[i+1:]...)
				i--
			}
			if !strings.Contains(args[i], "=") {
				args[i] = args[i] + "="
			}
		}
		sort.Strings(args)

		argsString := strings.Join(args, "&")
		res.WriteString(argsString + "\n")
	} else {
		res.WriteString("\n")
	}
	res.WriteString("host:" + r.Host + "\n")
	res.WriteString("\n")
	res.WriteString("host" + "\n" + "UNSIGNED-PAYLOAD")
	canonicalRequestString = res.String()
	xTime = r.URL.Query().Get("X-Amz-Date")
	Credential := r.URL.Query().Get("X-Amz-Credential")
	Credentials := strings.Split(Credential, "/")
	data, region, service := Credentials[1], Credentials[2], Credentials[3]
	credentialScope = data + "/" + region + "/" + service + "/" + "aws4_request"
	return canonicalRequestString, xTime, credentialScope
}

func buildStringToSign(canonicalRequestString string, xTime string, credentialScope string) string {
	return strings.Join([]string{
		"AWS4-HMAC-SHA256",
		xTime,
		credentialScope,
		hex.EncodeToString(makeHash(sha256.New(), []byte(canonicalRequestString))),
	}, "\n")
}

func deriveKey(secret, service, region string, t []byte) []byte {
	hmacDate := HMACSHA256([]byte("AWS4"+secret), t)
	hmacRegion := HMACSHA256(hmacDate, []byte(region))
	hmacService := HMACSHA256(hmacRegion, []byte(service))
	return HMACSHA256(hmacService, []byte("aws4_request"))
}

func buildSignature(secret string, strToSign string, time string, region string, serviceName string) string {
	key := deriveKey(secret, serviceName, region, []byte(time))
	return hex.EncodeToString(HMACSHA256(key, []byte(strToSign)))
}

func CalculateS3V4Signature(secret string, r *http.Request, region string, serviceName string) string {
	canonicalRequestString, xTime, credentialScope := GetCanonicalString(r)
	//log.Println("CalculateSignature()::canonicalRequestString:\n", canonicalRequestString, "\n ")
	Logger.Info("CalculateSignature()", zap.Any("canonicalRequestString", canonicalRequestString))
	stringToSign := buildStringToSign(canonicalRequestString, xTime, credentialScope)
	signingSignature := buildSignature(secret, stringToSign, xTime[0:timeLen], region, serviceName)
	return signingSignature
}

func CalculateAwsPreSignature(secret string, r *http.Request, serviceName string) string {
	canonicalRequestString, xTime, credentialScope := GetPreCanonicalString(r)
	Credential := r.URL.Query().Get("X-Amz-Credential")
	Credentials := strings.Split(Credential, "/")
	data, region, service := Credentials[1], Credentials[2], Credentials[3]
	signingkey := deriveKey(secret, service, region, []byte(data))
	strToSign := buildStringToSign(canonicalRequestString, xTime, credentialScope)
	return hex.EncodeToString(HMACSHA256(signingkey, []byte(strToSign)))
}

func CalculatePayloadHash(r *http.Request) (string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Logger.Error("CalculatePayloadHash()", zap.Any("err", err))
		return "", err
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return hex.EncodeToString(makeHash(sha256.New(), []byte(bodyBytes))), nil
}
