package main

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

const secret = "abc123abc123abc123"

var secretKey []byte

func NewURLSigner() {
	secretKey = []byte(secret)
}

func GenerateTokenFromString(data string) string {
	var urlToSign string

	s := goalone.New(secretKey, goalone.Timestamp)
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := s.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	return token
}

func VerifyToken(token string) bool {
	s := goalone.New(secretKey, goalone.Timestamp)
	_, err := s.Unsign([]byte(token))

	if err != nil {

		return false
	}

	return true

}

func Expired(token string, minutesUntilExpire int) bool {
	s := goalone.New(secretKey, goalone.Timestamp)
	ts := s.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
