package duden

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"golang.org/x/text/unicode/norm"

	"github.com/abdusco/alman/internal/http"
)

type fetcher interface {
	FetchHTML(ctx context.Context, url string) (string, error)
}

type Duden struct {
	fetcher fetcher
}

func New() *Duden {
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

var ErrNotFound = errors.New("not found")

func (d Duden) Find(ctx context.Context, word string) (Entry, error) {
	w := d.normalizeWord(word)
	log.Debug("searching duden", "word", word, "normalized", w)

	p := pool.NewWithResults[Entry]().WithContext(ctx)

	p.Go(func(ctx context.Context) (Entry, error) {
		return d.findUsingSearch(ctx, word)
	})

	p.Go(func(ctx context.Context) (Entry, error) {
		return d.findUsingURL(ctx, word)
	})

	results, _ := p.Wait()

	if len(results) == 0 {
		return Entry{}, ErrNotFound
	}

	return results[0], nil
}

func (d Duden) findUsingSearch(ctx context.Context, word string) (Entry, error) {
	searchURL := fmt.Sprintf("https://www.duden.de/suchen/dudenonline/%s", d.normalizeWord(word))
	html, err := d.fetcher.FetchHTML(ctx, searchURL)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to fetch html: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Entry{}, fmt.Errorf("failed to parse html: %w", err)
	}

	var links []string
	doc.Find("a[href*='/rechtschreibung/']").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		links = append(links, fmt.Sprintf("https://www.duden.de%s", href))
	})

	if len(links) == 0 {
		return Entry{}, ErrNotFound
	}

	u, _ := url.Parse(links[0])
	parts := strings.Split(u.Path, "/rechtschreibung/")
	foundWord, _ := lo.Last(parts)

	return d.findUsingURL(ctx, foundWord)
}

func (d Duden) findUsingURL(ctx context.Context, word string) (Entry, error) {
	u := fmt.Sprintf("https://www.duden.de/rechtschreibung/%s", d.normalizeWord(word))
	html, err := d.fetcher.FetchHTML(ctx, u)
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

	// TODO: handle sub-definitions: https://www.duden.de/rechtschreibung/rechnen
	doc.Find("#bedeutung").Each(func(_ int, el *goquery.Selection) {
		meaning := d.cleanText(el.Find(".division__header + p").Text())
		var examples []string
		el.Find(".note__list > li").Each(func(_ int, el *goquery.Selection) {
			examples = append(examples, d.cleanText(el.Text()))
		})

		entry.Definitions = append(entry.Definitions, Definition{
			Definition: meaning,
			Examples:   examples,
		})
	})
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
	"Ü":      "Ue",
	"Ä":      "Ae",
	"Ö":      "OE",
	"ü":      "ue",
	"ä":      "ae",
	"ö":      "oe",
	"ß":      "ss",
	"\u00ad": "",
}

func (d Duden) normalizeWord(word string) string {
	word = strings.TrimSpace(word)
	word = norm.NFC.String(word) // normalize umlauts with 2 chars into 1 char. e.g. ü -> ü
	for find, replace := range replacements {
		word = strings.ReplaceAll(word, find, replace)
	}
	return word
}
