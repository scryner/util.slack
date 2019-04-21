package msgfmt

import (
	"encoding/json"
)

type ButtonElement struct {
	Text Text
	URL  string
}

func (btn ButtonElement) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type": "button",
		"text": btn.Text,
		"url":  btn.URL,
	}

	return json.Marshal(m)
}

func (ButtonElement) elementAble() {}
