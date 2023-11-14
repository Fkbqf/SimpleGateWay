package public

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

func GenSaltPassword(salt, password string) string {
	s1 := sha256.New()
	s1.Write([]byte(password))
	str1 := fmt.Sprintf("%x", s1.Sum(nil))

	s2 := sha256.New()
	s2.Write([]byte(str1 + password))
	return fmt.Sprintf("%x", s2.Sum(nil))

}
func MD5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func OBj2Json(s interface{}) string {
	marshal, _ := json.Marshal(s)
	return string(marshal)
}
func InStringSlice(slice []string, str string) bool {
	for _, item := range slice {
		if str == item {
			return true
		}
	}
	return false
}
