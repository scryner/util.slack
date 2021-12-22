package block

import (
	"encoding/json"
)

type Input struct {
	Label   PlainText
	Element InputElement
}

func (i Input) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":    "input",
		"element": i.Element,
		"label":   i.Label,
	}

	return json.Marshal(m)
}

func (Input) blockAble() {}

type InputElement interface {
	json.Marshaler
	inputElementAble()
}

type PlainTextInput struct {
	Multiline    bool
	ActionId     string
	PlaceHolder  PlainText
	InitialValue string
	FocusOnLoad  bool
}

func (p PlainTextInput) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":          "plain_text_input",
		"multiline":     p.Multiline,
		"action_id":     p.ActionId,
		"focus_on_load": p.FocusOnLoad,
	}

	if p.PlaceHolder.Text != "" {
		m["placeholder"] = p.PlaceHolder
	}

	if p.InitialValue != "" {
		m["initial_value"] = p.InitialValue
	}

	return json.Marshal(m)
}

func (p PlainTextInput) inputElementAble() {}
