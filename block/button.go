package block

import (
	"encoding/json"
)

type Button struct {
	Text     PlainText
	Value    string
	ActionId string
}

func (btn Button) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":  "button",
		"text":  btn.Text,
		"value": btn.Value,
	}

	if btn.ActionId != "" {
		m["action_id"] = btn.ActionId
	}

	return json.Marshal(m)
}

func (Button) actionsElementAble()   {}
func (Button) sectionAccessoryAble() {}
