package common

import (
	"encoding/json"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/labstack/gommon/log"
)

// ToJSONString converts the given object to JSON string.
func ToJSONString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Errorf("[ToJSONString] Error when trying to marshal the object: %v error: %v", v, err)
		return constant.EmptyStr, err
	}
	return string(b), nil
}
