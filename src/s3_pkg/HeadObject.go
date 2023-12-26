// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Zijun Hu
// ///////////////////////////////////////
package s3_pkg

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

func (pkg *S3Pkg) HeadObject(
	bucketName, objectName string,
	urlQuery url.Values,
) (resMp map[string]string, err error) {
	err = pkg.GetS3Client()
	if err != nil {
		Logger.Info("HeadObject", zap.Any("error", err))
		return nil, err
	}

	s3HeadObjInput := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	if partNumber := GetQueryString(urlQuery, "partNumber"); partNumber != "" {
		intPartNumber, _ := strconv.Atoi(partNumber)
		s3HeadObjInput.PartNumber = int32(intPartNumber)
	}

	if versionid := GetQueryString(urlQuery, "versionId"); versionid != "" {
		s3HeadObjInput.VersionId = &versionid
	}

	result, err := pkg.S3Client.HeadObject(context.TODO(), s3HeadObjInput)
	if err != nil {
		Logger.Error("S3Pkg.HeadObject", zap.Any("bucketName", bucketName),
			zap.Any("error", err))
		return nil, err
	}

	resMp = make(map[string]string)

	if !result.DeleteMarker {
		resMp["x-amz-delete-marker"] = strconv.FormatBool(result.DeleteMarker)
	}
	if result.AcceptRanges != nil {
		resMp["Accept-Ranges"] = *result.AcceptRanges
	}
	if result.Expiration != nil {
		resMp["x-amz-expiration"] = *result.Expiration
	}
	if result.Restore != nil {
		resMp["x-amz-restore"] = *result.Restore
	}
	if result.ArchiveStatus != "" {
		resMp["x-amz-archive-status"] = string(result.ArchiveStatus)
	}

	if result.LastModified != nil {
		resMp["Last-Modified"] =
			string(result.LastModified.Format("Mon, 02 Jan 2006 03:04:05 GMT"))
	}

	resMp["Content-Length"] = strconv.FormatInt(result.ContentLength, 10)

	if result.ChecksumCRC32 != nil {
		resMp["x-amz-checksum-crc32"] = *result.ChecksumCRC32
	}
	if result.ChecksumCRC32C != nil {
		resMp["x-amz-checksum-crc32c"] = *result.ChecksumCRC32C
	}
	if result.ChecksumSHA1 != nil {
		resMp["x-amz-checksum-sha1"] = *result.ChecksumSHA1
	}
	if result.ChecksumSHA256 != nil {
		resMp["x-amz-checksum-sha256"] = *result.ChecksumSHA256
	}

	if result.ETag != nil {
		resMp["ETag"] = *result.ETag
	}
	if result.MissingMeta != 0 {
		resMp["x-amz-missing-meta"] = strconv.FormatInt(int64(result.MissingMeta), 10)
	}
	if result.VersionId != nil {
		resMp["x-amz-version-id"] = *result.VersionId
	}

	if result.CacheControl != nil {
		resMp["Cache-Control"] = *result.CacheControl
	}
	if result.ContentDisposition != nil {
		resMp["Content-Disposition"] = *result.ContentDisposition
	}
	if result.ContentEncoding != nil {
		resMp["Content-Encoding"] = *result.ContentEncoding
	}
	if result.ContentLanguage != nil {
		resMp["Content-Language"] = *result.ContentLanguage
	}
	if result.ContentType != nil {
		resMp["Content-Type"] = *result.ContentType
	}

	if result.Expires != nil {
		resMp["Expires"] = result.Expires.String()
	}
	if result.WebsiteRedirectLocation != nil {
		resMp["x-amz-website-redirect-location"] = *result.WebsiteRedirectLocation
	}

	if result.StorageClass != "" {
		resMp["x-amz-storage-class"] = string(result.StorageClass)
	}
	if result.RequestCharged != "" {
		resMp["x-amz-request-charged"] = string(result.RequestCharged)
	}
	if result.ReplicationStatus != "" {
		resMp["x-amz-replication-status"] = string(result.ReplicationStatus)
	}
	if result.PartsCount != 0 {
		resMp["x-amz-mp-parts-count"] = strconv.FormatInt(int64(result.PartsCount), 10)
	}

	if result.ObjectLockMode != "" {
		resMp["x-amz-object-lock-mode"] = string(result.ObjectLockMode)
	}
	if result.ObjectLockRetainUntilDate != nil {
		resMp["x-amz-object-lock-retain-until-date"] = result.ObjectLockRetainUntilDate.String()
	}
	if result.ObjectLockLegalHoldStatus != "" {
		resMp["x-amz-object-lock-legal-hold"] = string(result.ObjectLockLegalHoldStatus)
	}

	if result.Metadata != nil {
		for key, val := range result.Metadata {
			lowKey := strings.ToLower(key)
			resMp["X-Amz-Meta-"+lowKey] = val
		}
	}
	Logger.Info("S3Pkg.HeadObject(): ok.", zap.Any("result", resMp))

	return resMp, nil
}
