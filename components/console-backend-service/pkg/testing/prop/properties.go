package prop

import (
	"math/rand"
)

func String(length int) string {
	return StringWithCharset(length, Alphanumeric)
}

func StringWithCharset(length int, charset Charset) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}


func Labels(n int) map[string]string {
	result := make(map[string]string)
	for i := 0; i < n ; i++ {
		result[String(4)] = String(4)
	}
	return result
}

func OneOfString(vals ...string) string {
	return vals[rand.Intn(len(vals))]
}
