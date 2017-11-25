package template

import (
	"testing"

	"github.com/matryer/is"
)

func setupParser(t *testing.T) (*is.I, *Parser) {
	return is.New(t), &Parser{}
}

func TestParserAttemptsToProcessJSON(t *testing.T) {
	is, p := setupParser(t)

	tmpl := `{ "message": "{{.JSON.something}}" }`

	data, err := p.Parse(tmpl, []byte(`{ "something": "else" }`))

	is.NoErr(err)                                   // expected no error to be returned
	is.Equal(`{ "message": "else" }`, string(data)) // expected to correctly process data
}

func TestParserAttemptsToProcessRaw(t *testing.T) {
	is, p := setupParser(t)

	tmpl := `{ "message": "{{printf "%s" .Raw}}" }`

	data, err := p.Parse(tmpl, []byte("else"))

	is.NoErr(err)                                   // expected no error to be returned
	is.Equal(`{ "message": "else" }`, string(data)) // expected to correctly process data
}

func TestParserReturnsErrorWithBadTemplate(t *testing.T) {
	is, p := setupParser(t)

	tmpl := `{ "message": "{{.something}`

	_, err := p.Parse(tmpl, nil)

	is.True(err != nil) // expected error to be returned
}
