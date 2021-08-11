package block

import (
	"encoding/json"
)

type Actions struct {
	Elements []ActionsElement
}

func (a Actions) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":     "actions",
		"elements": a.Elements,
	})
}

func (Actions) blockAble() {}

type ActionsElement interface {
	json.Marshaler
	actionsElementAble()
}

type CheckBoxesActionOption struct {
	Text        PlainText `json:"text"`
	Description PlainText `json:"description"`
	Value       string    `json:"value"`
}

type CheckBoxesAction struct {
	Options  []CheckBoxesActionOption
	ActionId string
}

func (cbs CheckBoxesAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":      "checkboxes",
		"options":   cbs.Options,
		"action_id": cbs.ActionId,
	})
}

func (CheckBoxesAction) actionsElementAble() {}
