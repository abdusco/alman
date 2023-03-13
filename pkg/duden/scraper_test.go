package duden

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeFetcher string

func (f fakeFetcher) FetchHTML(context.Context, string) (string, error) {
	return string(f), nil
}

func TestDuden_findUsingURL(t *testing.T) {
	tests := []struct {
		name      string
		word      string
		html      fakeFetcher
		assertRes func(t *testing.T, res Entry, err error)
	}{
		{
			name: "happy path",
			word: "betreuen",
			html: sampleHtml,
			assertRes: func(t *testing.T, res Entry, err error) {
				expected := Entry{
					Word: "betreuen",
					Definitions: []Definition{
						{
							Definition: "vorübergehend in seiner Obhut haben, in Obhut nehmen; für jemanden, etwas sorgen",
							Examples: []string{
								"Kinder, alte Leute, Tiere betreuen",
								"eine Reiseleiterin betreut die Gruppe",
								"die Sportler werden von einem Trainer betreut",
								"betreutes (ein mit einer Betreuung der betreffenden Person[en] verbundenes) Wohnen",
							},
						}, {
							Definition: "ein Sachgebiet o. Ä. fortlaufend bearbeiten; die Verantwortung für den Ablauf von etwas haben",
							Examples: []string{
								"eine Abteilung, ein Arbeitsgebiet betreuen",
								"sie betreut das Projekt zur Sanierung der Altbauten",
							},
						}},
				}
				assert.NoError(t, err)
				assert.Equal(t, expected, res)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Duden{
				fetcher: tt.html,
			}
			actual, err := d.findUsingURL(context.Background(), tt.word)
			tt.assertRes(t, actual, err)
		})
	}
}

const sampleHtml = `
<div>
<h1 class="lemma__title lemma__title--short">
          <span class="lemma__main">be­treu­en</span>
        </h1>
<div class="division " id="bedeutungen">
  <header class="division__header">
    
          <h2 class="division__title">Bedeutungen (2)</h2>
              <small class="division__info">
        <a class="division__info_icon" target="_blank" href="/hilfe/bedeutungen">ⓘ</a>
      </small>
        
  </header>
    
      
      <ol class="enumeration" type="a"><li class="enumeration__item" id="Bedeutung-a">
          <div class="enumeration__text">vorübergehend in seiner Obhut haben, in Obhut nehmen; für jemanden, etwas sorgen</div>
          <dl class="note"><dt class="note__title">Beispiele</dt>
            <dd>
              <ul class="note__list"><li>Kinder, alte Leute, Tiere betreuen</li>
                <li>eine Reiseleiterin betreut die Gruppe</li>
                <li>die Sportler werden von einem Trainer betreut</li>
                <li>betreutes <i>(ein mit einer Betreuung der betreffenden Person[en] verbundenes) </i>Wohnen</li>
              </ul></dd>
          </dl></li>
        <li class="enumeration__item" id="Bedeutung-b">
          <div class="enumeration__text">ein Sachgebiet o.&nbsp;Ä. fortlaufend bearbeiten; die Verantwortung für den Ablauf von etwas haben</div>
          <dl class="note"><dt class="note__title">Beispiele</dt>
            <dd>
              <ul class="note__list"><li>eine Abteilung, ein Arbeitsgebiet betreuen</li>
                <li>sie betreut das Projekt zur Sanierung der Altbauten</li>
              </ul></dd>
          </dl></li>
      </ol>
  
</div>
</div>
`
