package msgfmt

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Blocks []Block

func (bs Blocks) MarshalJSON() ([]byte, error) {
	var marshaledBlocks []string

	for _, b := range bs {
		marshaled, err := json.Marshal(b)
		if err != nil {
			return nil, err
		}

		marshaledBlocks = append(marshaledBlocks, string(marshaled))
	}

	j := fmt.Sprintf(`{"blocks":[%s]}`, strings.Join(marshaledBlocks, ","))

	return []byte(j), nil
}

func (bs Blocks) sendAble() {}

type Block interface{
	json.Marshaler
	blockAble()
}

type Element interface {
	json.Marshaler
	elementAble()
}
