<small> [简体中文](README_zh.md) | English </small>

# Oceanpass | [Documentation](docs/)
[![GitHub license](https://img.shields.io/badge/license-apache--2--Clause-brightgreen.svg)](./LICENSE)

一个多云的对象存储接入框架。通过简单的配置，用户能将公有云对象存储和私有云对象存储接入到oceanpass中。

# 用户指导
将代码仓库clone到本地
```bash
git clone git@github.com:sfireworks/Oceanpass.git
cd oceanpss
```
## 如何配置
### 配置Access key和secrest config
```bash
vim credentials
# replace XXXX with your own object storage access key and secret
[default]
aws_access_key_id = XXXXXXXXXXXXXXXX
aws_secret_access_key = XXXXXXXXXXXXXXXXXXXXX
```

### 配置Endpoint
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

## 启动oceanpass服务
在本地运行

```bash
go build -o oceanpass_http_main  src/oceanpass_http_main.go
./oceanpass_http_main
```

以容器方式运行

```bash
docker build -t oceanpass:latest -f Dockerfile .
docker run -d oceanpass:latest
```

## 性能测试
由于我们在原本的对象存储系统之上增加了一层，所以oceanpass增加了月5%的访问延迟。

## 兼容性
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

## 设计
[Detailed design document](docs/oceanpass-design.zh.md)

## 依赖
* Amazon S3 SDK
* Aliyun OSS SDK

## 即将到来
- CI和测试覆盖
- 支持使用块存储和文件存储来缓存数据
- 大文件下载性能优化
- 支持更多的对象存储服务提供商

## 联系我们
  * Issue: [this link](https://github.com/sfireworks/oceanpass/issues)
  * Email: sfireworks@qq.com

## 贡献者
  * [sfireworks](https://github.com/sfireworks)
  * [Eric](https://github.com/rhinouser0)


## License
- [Apache 2.0](LICENSE)