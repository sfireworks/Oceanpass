// ///////////////////////////////////////
// 2022 SHAILab Storage all rights reserved
// Author: Chao Qin
// ///////////////////////////////////////
package common

import (
	"bytes"
	"encoding/xml"
	"net/url"
)

func GetQueryString(vars url.Values, key string) string {
	if len(vars[key]) > 0 {
		return vars[key][0]
	}
	return ""
}

func MarshalIndent(input interface{}) ([]byte, error) {
	resXML, err := xml.MarshalIndent(input, " ", " ")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write([]byte(xml.Header))
	buffer.Write(resXML)
	buffer.Write([]byte("\n\r"))
	res := buffer.Bytes()

	return res, nil
}
