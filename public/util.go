package public

import (
	"crypto/sha1"
	"fmt"
)

func GenerateSaltPassword(salt, password string) string {
	s1 := sha1.New()
	s1.Write([]byte(password))
	str1 := fmt.Sprintf("%x", s1.Sum(nil))

	s2 := sha1.New()
	s2.Write([]byte(str1 + salt))
	return fmt.Sprintf("%x", s2.Sum(nil))
}
