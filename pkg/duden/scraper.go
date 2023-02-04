package duden

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

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
	Word     string    `json:"word"`
	Meanings []Meaning `json:"meanings"`
}

type Meaning struct {
	Meaning  string   `json:"meaning"`
	Examples []string `json:"examples"`
}

const entryTemplate = `
# {{.Word}}
{{- range .Meanings }}

## {{ .Meaning }}
{{- range .Examples }}
- {{ . }}
{{- end }}
{{- end -}}
`

func (e DictionaryEntry) String() string {
	t := template.Must(template.New("a").Parse(entryTemplate))
	var b bytes.Buffer
	t.Execute(&b, e)
	return strings.TrimSpace(b.String())
}

func (d Duden) Find(word string) (DictionaryEntry, error) {
	return d.findUsingURL(word)
}

func (d Duden) findUsingURL(word string) (DictionaryEntry, error) {
	word = d.normalizeWord(word)
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

var umlautMap = map[string]string{
	"\u00dc": "UE",
	"\u00c4": "AE",
	"\u00d6": "OE",
	"\u00fc": "ue",
	"\u00e4": "ae",
	"\u00f6": "oe",
	"\u00df": "ss",
	"\u00ad": "",
}

func (d Duden) normalizeWord(word string) string {
	for find, replace := range umlautMap {
		word = strings.ReplaceAll(word, find, replace)
	}
	return word
}
