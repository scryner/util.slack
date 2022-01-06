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

func (PlainTextInput) inputElementAble() {}

type SelectOption struct {
	Text  PlainText `json:"text"`
	Value string    `json:"value"`
}

type StaticSelect struct {
	Placeholder PlainText
	Options     []SelectOption
	ActionId    string
}

func (s StaticSelect) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":        "static_select",
		"placeholder": s.Placeholder,
		"options":     s.Options,
		"action_id":   s.ActionId,
	}

	return json.Marshal(m)
}

func (StaticSelect) inputElementAble() {}
