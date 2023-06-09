package dwds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"

	"github.com/abdusco/alman/internal/http"
)

type fetcher interface {
	FetchHTML(ctx context.Context, url string) (string, error)
}

// Dwds stands for "Digitales Wörterbuch der deutschen Sprache"
type Dwds struct {
	fetcher fetcher
}

func New() Dwds {
	return Dwds{fetcher: http.ReqFetcher}
}

type Entry struct {
	Word        string       `json:"word"`
	Definitions []Definition `json:"definitions"`
	Usages      []string     `json:"usages"`
}

const entryTemplate = `
# {{.Word}}
{{- range .Definitions }}

## {{ .Definition }}
{{- range .Examples }}
- {{ . }}
{{- end }}
{{- end -}}

{{ if .Usages }}

---

{{ range .Usages -}}
- {{ . }}
{{ end -}}
{{- end -}}
`

func (e Entry) String() string {
	t := template.Must(template.New("a").Parse(entryTemplate))
	var b bytes.Buffer
	_ = t.Execute(&b, e)
	return strings.TrimSpace(b.String())
}

type Definition struct {
	Definition string   `json:"definition"`
	Examples   []string `json:"examples"`
}

var ErrNotFound = errors.New("not found")

func (d Dwds) Find(ctx context.Context, word string) (Entry, error) {
	u := "https://www.dwds.de/?q=" + url.QueryEscape(word)
	html, err := d.fetcher.FetchHTML(ctx, u)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to fetch html: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return Entry{}, fmt.Errorf("failed to parse html: %w", err)
	}

	if errText := strings.TrimSpace(doc.Find(".bg-danger").Text()); regexp.MustCompile(`(?smi)nicht.+vorhanden`).MatchString(errText) {
		return Entry{}, ErrNotFound
	}

	doc.RemoveFiltered(".dwdswb-fundstelle")

	entry := Entry{
		Word: strings.TrimSpace(doc.Find(".dwdswb-ft-lemmaansatz").Text()),
	}

	doc.Find(".dwdswb-lesart-def").Each(func(_ int, el *goquery.Selection) {
		var examples []string
		el.NextAllFiltered(".dwdswb-verwendungsbeispiele").Each(func(_ int, el *goquery.Selection) {
			el.Find(".dwdswb-kompetenzbeispiel, .dwdswb-beleg .dwdswb-belegtext").Each(func(_ int, el *goquery.Selection) {
				examples = append(examples, strings.TrimSpace(el.Text()))
			})
		})

		var definition string
		for _, fn := range []func() string{
			func() string {
				return strings.TrimSpace(el.Find(".dwdswb-definition").Text())
			},
			func() string {
				ref := el.Find(".dwdswb-verweis")
				if content, ok := ref.Attr("data-content"); ok {
					d, _ := goquery.NewDocumentFromReader(strings.NewReader(content))
					return fmt.Sprintf("%s (%s)", ref.Text(), d.Text())
				}
				return ""
			},
		} {
			if definition = fn(); definition != "" {
				break
			}
		}

		if definition == "" {
			return
		}

		entry.Definitions = append(entry.Definitions, Definition{
			Definition: definition,
			Examples:   examples,
		})
	})

	doc.Find(`[data-content-piece="Verwendungsbeispiele"] .dwdswb-belegtext`).Each(func(_ int, el *goquery.Selection) {
		entry.Usages = append(entry.Usages, strings.TrimSpace(el.Text()))
	})

	return entry, nil
}
