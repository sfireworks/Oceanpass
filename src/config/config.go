// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package config

import (
	"bufio"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Config struct {
	XMLName                xml.Name       `xml:"server_config"`
	AddrConfig             AddressConfig  `xml:"address_config"`
	EndpointConfig         EndpointConfig `xml:"endpoint_config"`
	ServerMode             string         `xml:"server_mode"`
	ServiceName            string         `xml:"service_name"`
	StsDefaultDurationSecs string         `xml:"sts_default_duration_seconds"`

	SignatureDisabled       bool `xml:"signature_all_disabled"`
	SignatureForOssReqAbled bool `xml:"signature_oss_req_abled"`
}

type AddressConfig struct {
	ConfigFlag string `xml:"config_name,attr"`
	IpAddr     string `xml:"ip_addr"`
	HttpPort   string `xml:"http_port"`
	GrpcPort   string `xml:"grpc_port"`
}

type EndpointConfig struct {
	ConfigFlag               string      `xml:"config_name,attr"`
	CloudProviderForEndpoint string      `xml:"cloud_provider_for_endpoint"`
	Endpoints                []Endpoints `xml:"endpoint"`
}

type Endpoints struct {
	ConfigFlag  string `xml:"config_name,attr"`
	EndpointUrl string `xml:"endpoint_url"`
}

func (cfg *Config) LoadXMLConfig(config_path string) {
	if config_path == "" {
		config_path = "server_config_backup.xml"
	}
	xmlFile, err := os.Open(config_path)
	if err != nil {
		log.Fatalf("Error opening XML file! path: %v\n", config_path)
	}
	defer xmlFile.Close()

	xmlData, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatalln("Error reading XML data:", err)
	}

	xml.Unmarshal(xmlData, &cfg)
	log.Println("XMLName : ", cfg.XMLName)
	log.Printf("XMLName : %+v ", cfg)
	log.Println()
}

func GetConfigFromFile(path string) map[string]string {
	config := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config
}
