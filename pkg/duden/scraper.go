package duden

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"

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

type Entry struct {
	Word        string       `json:"word"`
	Definitions []Definition `json:"definitions"`
}

type Definition struct {
	Definition string   `json:"definition"`
	Examples   []string `json:"examples"`
}

const entryTemplate = `
# {{.Word}}
{{- range .Definitions }}

## {{ .Definition }}
{{- range .Examples }}
- {{ . }}
{{- end }}
{{- end -}}
`

func (e Entry) String() string {
	t := template.Must(template.New("a").Parse(entryTemplate))
	var b bytes.Buffer
	_ = t.Execute(&b, e)
	return strings.TrimSpace(b.String())
}

func (d Duden) Find(word string) (Entry, error) {
	log.Debug().Str("word", word).Msg("searching duden")
	return d.findUsingURL(word)
}

func (d Duden) findUsingURL(word string) (Entry, error) {
	word = d.normalizeWord(word)
	url := fmt.Sprintf("https://www.duden.de/rechtschreibung/%s", word)
	html, err := d.fetcher.FetchHTML(url)
	if err != nil {
		if errors.Is(err, http.ErrNotFound) {
		}
		return Entry{}, fmt.Errorf("failed to fetch html: %w", err)
	}
	return d.parseEntry(html)
}

func (d Duden) parseEntry(html string) (Entry, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Entry{}, fmt.Errorf("failed to parse html: %w", err)
	}

	var entry Entry

	title := d.cleanText(doc.Find(".lemma__title").Text())
	entry.Word = title

	doc.Find("#bedeutungen .enumeration__item").Each(func(_ int, el *goquery.Selection) {
		meaning := d.cleanText(el.Find(".enumeration__text").Text())
		var examples []string
		el.Find(".note__list > li").Each(func(_ int, el *goquery.Selection) {
			examples = append(examples, d.cleanText(el.Text()))
		})

		entry.Definitions = append(entry.Definitions, Definition{
			Definition: meaning,
			Examples:   examples,
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

var replacements = map[string]string{
	"Ü":      "UE",
	"Ä":      "AE",
	"Ö":      "OE",
	"ü":      "ue",
	"ä":      "ae",
	"ö":      "oe",
	"ß":      "ss",
	"\u00ad": "",
}

func (d Duden) normalizeWord(word string) string {
	word = strings.TrimSpace(word)
	for find, replace := range replacements {
		word = strings.ReplaceAll(word, find, replace)
	}
	return word
}
