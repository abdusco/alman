package http

import (
	"time"

	"github.com/imroc/req/v3"
)

var DefaultClient = req.NewClient().
	SetTimeout(time.Second * 10).
	DisableAutoReadResponse().
	EnableInsecureSkipVerify()
