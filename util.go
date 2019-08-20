package apiintegration

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

//ToString convert type (int, int64, float32, float64, byte, and []bytes) to string
//Parameter p is optional and only used in converting float
func ToString(n interface{}, p ...int) string {
	var t string

	switch n.(type) {
	case bool:
		t = strconv.FormatBool(n.(bool))
	case int:
		t = strconv.Itoa(n.(int))
	case int64:
		t = strconv.FormatInt(n.(int64), 10)
	case float32:
		if len(p) > 0 {
			t = strconv.FormatFloat(float64(n.(float32)), 'f', p[0], 64)
		} else {
			t = strconv.FormatFloat(float64(n.(float32)), 'f', -1, 64)
		}
	case float64:
		if len(p) > 0 {
			t = strconv.FormatFloat(n.(float64), 'f', p[0], 64)
		} else {
			t = strconv.FormatFloat(n.(float64), 'f', -1, 64)
		}
	case byte:
		t = string(n.(byte))
	case []byte:
		t = string(n.([]byte))
	case string:
		t = n.(string)
	}

	return t
}

//EncodeMD5 : encrypt to MD5 input string, output to string
func EncodeMD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
