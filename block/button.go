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
	return json.Marshal(map[string]interface{}{
		"type":      "button",
		"text":      btn.Text,
		"value":     btn.Value,
		"action_id": btn.ActionId,
	})
}

func (Button) actionsElementAble()   {}
func (Button) sectionAccessoryAble() {}
