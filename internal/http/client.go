package http

import (
	"os"
	"time"

	"github.com/imroc/req/v3"
)

var DefaultClient = req.NewClient().
	SetTimeout(time.Second * 10).
	SetUserAgent(envOrDefault("USER_AGENT", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")).
	DisableAutoReadResponse().
	EnableInsecureSkipVerify()

func envOrDefault(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		value = defaultValue
	}
	return value
}
