package block

import (
	"encoding/hex"
	"strings"
)

type Color string

const (
	GoodColor    Color = "good"
	WarningColor Color = "warning"
	DangerColor  Color = "danger"
)

func (c Color) String() string {
	return string(c)
}

func RGB(r, g, b uint8) Color {
	return Color(strings.ToUpper(hex.EncodeToString([]byte{r, g, b})))
}

type Attachment struct {
	Blocks []Block `json:"blocks,omitempty"`
	Color  Color   `json:"color,omitempty"`
}
