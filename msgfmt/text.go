package msgfmt

import (
	"encoding/json"
)

type Text interface {
	text() string
	Message
}

type PlainText struct {
	Text  string
	Emoji bool
}

func (t PlainText) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":  "plain_text",
		"text":  t.Text,
		"emoji": t.Emoji,
	}

	return json.Marshal(m)
}

func (t PlainText) text() string {
	return t.Text
}

func (t PlainText) sendAble()             {}
func (t PlainText) sectionAccessoryAble() {}

type MarkdownText struct {
	Text string
}

func (t MarkdownText) text() string {
	return t.Text
}

func (t MarkdownText) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type": "mrkdwn",
		"text": t.Text,
	}

	return json.Marshal(m)
}

func (t MarkdownText) sendAble()             {}
func (t MarkdownText) sectionAccessoryAble() {}
