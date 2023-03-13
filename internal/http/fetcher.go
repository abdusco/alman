package http

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/log"
)

type htmlFetcher struct {
}

var ErrNotFound = errors.New("page not found")

func (s htmlFetcher) FetchHTML(ctx context.Context, url string) (string, error) {
	log.Debug("fetching url", "url", url)
	res, err := DefaultClient.R().SetContext(ctx).Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}
	log.Debug("got response", "status_code", res.StatusCode, "url", res.Request.RawURL)
	if res.IsErrorState() {
		return "", ErrNotFound
	}
	return res.ToString()
}

var ReqFetcher = htmlFetcher{}
