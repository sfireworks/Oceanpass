// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package s3_pkg

type ListBucketsReq struct {
}

// TODO: finish it and change name to ListBuckets
func (pkg *S3Pkg) ListBuckets(
	endpoint string, bucketName string,
	paramsMap map[string]string) ([]byte, error) {

	var res []byte
	var err error
	return res, err
}
