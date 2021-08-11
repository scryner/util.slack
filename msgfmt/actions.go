package msgfmt

import (
	"encoding/json"
)

type Button struct {
	Text Text
	URL  string
}

func (btn Button) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type": "button",
		"text": btn.Text,
		"url":  btn.URL,
	}

	return json.Marshal(m)
}

func (Button) sectionAccessoryAble() {}
