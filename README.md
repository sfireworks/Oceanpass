<small> [简体中文](README_zh.md) | English </small>

# Oceanpass | [Documentation](docs/)
[![GitHub license](https://img.shields.io/badge/license-apache--2--Clause-brightgreen.svg)](./LICENSE)

A multi-cloud object storage intergration framework. With simple configurations, you can intergrate a public cloud object store or your private object storage system into oceanpass.

# Usage Guide
Clone the repo to your local enviroment with 
```bash
git clone xxx/oceanpass
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