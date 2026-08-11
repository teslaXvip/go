package http

import (
	"net/url"
	"strings"
)

func EncodeUrlParams(params map[string]string) string {
	sb := strings.Builder{}
	for k, v := range params {
		sb.WriteString(url.QueryEscape(k))
		sb.WriteString("=")
		sb.WriteString(url.QueryEscape(v))
		sb.WriteString("&")
	}
	if sb.Len() > 1 {
		return sb.String()[0 : sb.Len()-1]
	}
	return sb.String()
}

func ParseUrlParams(rawQuery string) map[string]string {
	params := make(map[string]string, 10)
	args := strings.Split(rawQuery, "&")
	for _, ele := range args {
		arr := strings.Split(ele, "=")
		if len(arr) == 2 {
			key, _ := url.QueryUnescape(arr[0])
			value, _ := url.QueryUnescape(arr[1])
			params[key] = value
		}
	}
	return params
}
