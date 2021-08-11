package block

import (
	"encoding/json"
)

type Message interface {
	json.Marshaler
	sendAble()
}
