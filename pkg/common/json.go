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

// FromJSONString converts the given JSON string to object.
func FromJSONString(s string, v interface{}) error {
	err := json.Unmarshal([]byte(s), v)
	if err != nil {
		log.Errorf("[FromJSONString] Error when trying to unmarshal the string: %s error: %v", s, err)
		return err
	}
	return nil
}
