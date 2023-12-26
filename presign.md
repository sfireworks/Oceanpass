create presigned URL
` ./ossutil64 sign oss://csun-meta-bkt-1/logging.json --timeout 3600 -e http://0.0.0.0:xxxx`

send Request
` curl `+ response of ` ./ossutil64 sign oss://csun-meta-bkt-1/logging.json --timeout 3600 -e http://0.0.0.0:xxxx`