package http

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

type htmlFetcher struct {
}

var ErrNotFound = errors.New("page not found")

func (s htmlFetcher) FetchHTML(url string) (string, error) {
	log.Debug().Str("url", url).Msg("fetching url")
	res, err := DefaultClient.R().Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}
	log.Debug().Int("status_code", res.StatusCode).Str("url", res.Request.RawURL).Msg("got response")
	if res.IsErrorState() {
		return "", ErrNotFound
	}
	return res.ToString()
}

var ReqFetcher = htmlFetcher{}
