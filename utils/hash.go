package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"reflect"
)

func hashProcess(h hash.Hash, obj interface{}, fields ...string) []byte {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if len(fields) != 0 {
		for _, fieldName := range fields {
			_, ok := t.FieldByName(fieldName)
			if ok {
				v := v.FieldByName(fieldName).Interface()
				h.Write([]byte(fmt.Sprintf("%v", v)))
			}
		}
	} else {
		str := ""
		for i := 0; i < v.NumField(); i++ {
			vfmt := fmt.Sprintf("%v", v.Field(i).Interface())
			str += vfmt
			h.Write([]byte(vfmt))
		}
	}

	sum := h.Sum(nil)
	return sum
}

func hashProcessHex(h hash.Hash, obj interface{}, fields ...string) string {
	sum := hashProcess(h, obj, fields...)
	return hex.EncodeToString(sum)
}

func hashProcessBase64(h hash.Hash, obj interface{}, fields ...string) string {
	sum := hashProcess(h, obj, fields...)
	return base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(sum)
}

func Hash(obj interface{}, fields ...string) string {
	h := sha256.New()
	return hashProcessHex(h, obj, fields...)
}

func Hmac(secret string, obj interface{}, fields ...string) string {
	h := hmac.New(sha256.New, []byte(secret))
	return hashProcessHex(h, obj, fields...)
}

func HmacBase64(secret string, obj interface{}, fields ...string) string {
	h := hmac.New(sha256.New, []byte(secret))
	return hashProcessBase64(h, obj, fields...)
}

func HashValue(v interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", v)))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func HmacValue(secret string, v interface{}) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%v", v)))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

func HashValsHex(vals ...string) string {
	h := sha256.New()
	for _, val := range vals {
		h.Write([]byte(val))
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}
