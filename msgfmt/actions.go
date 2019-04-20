package msgfmt

import (
	"encoding/json"
)

type ButtonElement struct {
	Text  Text
	Value string
}

func (btn ButtonElement) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":  "button",
		"text":  btn.Text,
		"value": btn.Value,
	}

	return json.Marshal(m)
}

func (ButtonElement) elementAble() {}
