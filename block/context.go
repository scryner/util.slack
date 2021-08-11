package block

import (
	"encoding/json"
)

type Context struct {
	Elements []ContextElement
}

func (ctx Context) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":     "context",
		"elements": ctx.Elements,
	}

	return json.Marshal(m)
}

func (Context) blockAble() {}

type ContextElement interface {
	json.Marshaler
	contextElementAble()
}
