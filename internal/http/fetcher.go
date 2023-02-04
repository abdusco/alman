package http

import (
	"errors"
	"fmt"
)

type htmlFetcher struct {
}

var ErrNotFound = errors.New("page not found")

func (s htmlFetcher) FetchHTML(url string) (string, error) {
	res, err := DefaultClient.R().Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}
	if res.IsErrorState() {
		return "", ErrNotFound
	}
	return res.ToString()
}

var ReqFetcher = htmlFetcher{}
