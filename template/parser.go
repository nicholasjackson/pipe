package template

import (
	"bytes"
	"encoding/json"
	gotemplate "text/template"
)

// parserData defines the internal data structure for the parser
type parserData struct {
	Raw  []byte
	JSON map[string]interface{}
}

// Parser defines a new template parser
type Parser struct{}

// Parse attempts to parse the template with the given data
func (p *Parser) Parse(template string, data []byte) ([]byte, error) {
	pd := parserData{
		Raw: data,
	}

	// attempt to proecess json
	json.Unmarshal(data, &pd.JSON)

	tmpl, err := gotemplate.New("template").
		Funcs(gotemplate.FuncMap{
			"base64encode": base64encode,
			"base64decode": base64decode,
		}).
		Parse(template)

	if err != nil {
		return nil, err
	}

	out := bytes.NewBufferString("")
	err = tmpl.Execute(out, pd)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
