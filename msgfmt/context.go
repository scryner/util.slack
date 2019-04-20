package msgfmt

import (
	"encoding/json"
)

type Context struct {
	Elements []Element
}

func (ctx Context) MarshalJSON() ([]byte, error) {
	m := map[string]interface{} {
		"type": "context",
		"elements": ctx.Elements,
	}

	return json.Marshal(m)
}

func (Context) blockAble() {}
