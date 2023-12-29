<small> [简体中文](README_zh.md) | English </small>

# Oceanpass | [Documentation](docs/)
[![GitHub license](https://img.shields.io/badge/license-apache--2--Clause-brightgreen.svg)](./LICENSE)

A multi-cloud object storage intergration framework. With simple configurations, you can intergrate a public cloud object store or your private object storage system into oceanpass.

# Usage Guide
Clone the repo to your local enviroment with 
```bash
git clone git@github.com:sfireworks/Oceanpass.git
cd oceanpss
```
## How to configure
### Access key and secrest config
```bash
vim credentials
# replace XXXX with your own object storage access key and secret
[default]
aws_access_key_id = XXXXXXXXXXXXXXXX
aws_secret_access_key = XXXXXXXXXXXXXXXXXXXXX
```

### Endpoint config
```bash
vim conf/server_config.xml
# replace 9999 with your own service port
    <address_config config_name="address_config">
        <ip_addr>0.0.0.0</ip_addr>
        <http_port>9999</http_port>
    </address_config>

# add at least one endpoint of your object store and choose one as the storage provider
    <cloud_provider_for_endpoint>oss_cn-shanghai-internal</cloud_provider_for_endpoint>
        <endpoint config_name="oss_cn-shanghai-internal">
            <endpoint_url>https://oss-cn-shanghai-internal.aliyuncs.com</endpoint_url>
        </endpoint>
```

## Start oceanpass service
Run in the local environment

```bash
go build -o oceanpass_http_main  src/oceanpass_http_main.go
./oceanpass_http_main
```

Run as a docker

```bash
docker build -t oceanpass:latest -f Dockerfile .
docker run -d oceanpass:latest
```

## Performance test
Since we add a intergration layer ontop of object storage systems, oceanpass increases the access latency within 5%.

## Compatibility

| api list | awscli | oss java sdk |
| :-----| :----: | :----: |
| ListObjects | &#10004; | &#10004; |
| list_objects_v2 | &#10004; | &#10004; |
| head_object | &#10004; | &#10004; |
| get_object | &#10004; | &#10004; |
| put_object | &#10004; | &#10004; |
| copy_object | &#10004; | &#10004; |
| delete_object | &#10004; | &#10004; |
| delete_objects | &#10004; | &#10004; |
| create_multipart_upload | &#10004; | &#10004; |
| complete_multipart_upload | &#10004; | &#10004; |
| list_multipart_uploads | &#10004; | &#10004; |
| abort_multipart_upload | &#10004; | &#10004; |
| list_parts | &#10004; | &#10004; |
| upload_part | &#10004; | &#10004; |
| upload_part_copy | &#10004; | &#10004; |
| list_buckets | &#10004; | &#10004; |
| get_bucket_location | &#10004; | &#10004; |
| create_bucket | &#10004; | &#10004; |
| head_bucket | &#10004; | &#10004; |
| delete_bucket | &#10004; | &#10004; |
| get_bucket_acl | &#10004; | &#10004; |
| get_bucket_cors | &#10004; | &#10004; |
| put_bucket_cors | &#10004; | &#10004; |
| delete_bucket_cors | &#10004; | &#10004; |
| get_bucket_policy | &#10004; | &#10004; |
| put_bucket_policy | &#10004; | &#10004; |
| get_bucket_policy_status | &#10004; | &#10004; |
| get_bucket_logging | &#10004; | &#10004; |
| put_bucket_logging | &#10004; | &#10004; |

## Design
[Detailed design document](docs/oceanpass-design.zh.md)

## Dependency
* Amazon S3 SDK
* Aliyun OSS SDK

## Coming Soon
- CI and test coverage
- Support for cache object data in NAS or block storage
- Large file get optimization
- Support more cloud object storage providers  

## Contact Us
  * Issue: [this link](https://github.com/sfireworks/oceanpass/issues)
  * Email: sfireworks@qq.com

## Contributor
  * [sfireworks](https://github.com/sfireworks)
  * [Eric](https://github.com/rhinouser0)


## License
- [Apache 2.0](LICENSE)