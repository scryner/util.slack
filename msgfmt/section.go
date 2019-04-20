package msgfmt

import (
	"encoding/json"
)

type Section struct {
	Text Text
	Accessory Element
}

func (s Section) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type": "section",
		"text": s.Text,
	}

	if s.Accessory != nil {
		m["accessory"] = s.Accessory
	}

	return json.Marshal(m)
}

func (Section) blockAble() {}
