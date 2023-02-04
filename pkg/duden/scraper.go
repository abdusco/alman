package duden

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/abdusco/alman/internal/http"
)

type fetcher interface {
	FetchHTML(url string) (string, error)
}

type Duden struct {
	fetcher fetcher
}

func NewDuden() *Duden {
	return &Duden{fetcher: http.ReqFetcher}
}

type DictionaryEntry struct {
	Word     string
	Meanings []Meaning
}

type Meaning struct {
	Meaning  string
	Examples []string
}

func (d Duden) findUsingURL(word string) (DictionaryEntry, error) {
	url := fmt.Sprintf("https://www.duden.de/rechtschreibung/%s", word)
	html, err := d.fetcher.FetchHTML(url)
	if err != nil {
		if errors.Is(err, http.ErrNotFound) {
		}
		return DictionaryEntry{}, fmt.Errorf("failed to fetch html: %w", err)
	}
	return d.parseEntry(html)
}

func (d Duden) parseEntry(html string) (DictionaryEntry, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return DictionaryEntry{}, fmt.Errorf("failed to parse html: %w", err)
	}

	var entry DictionaryEntry

	title := d.cleanText(doc.Find(".lemma__title").Text())
	entry.Word = title

	doc.Find("#bedeutungen .enumeration__item").Each(func(_ int, el *goquery.Selection) {
		meaning := d.cleanText(el.Find(".enumeration__text").Text())
		var examples []string
		el.Find(".note__list > li").Each(func(_ int, el *goquery.Selection) {
			examples = append(examples, d.cleanText(el.Text()))
		})

		entry.Meanings = append(entry.Meanings, Meaning{
			Meaning:  meaning,
			Examples: examples,
		})
	})

	return entry, nil
}

func (d Duden) cleanText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\u00ad", "")
	text = strings.ReplaceAll(text, " ", " ")
	return text
}
