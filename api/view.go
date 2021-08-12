package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/scryner/util.slack/block"
	"github.com/scryner/util.slack/internal/crypto"
)

type PrivateMetadata []byte

type View struct {
	Type            string           `json:"type"`
	Title           block.PlainText  `json:"title"`
	Blocks          []block.Block    `json:"blocks"`
	Close           *block.PlainText `json:"close,omitempty"`
	Submit          *block.PlainText `json:"submit,omitempty"`
	PrivateMetadata PrivateMetadata  `json:"private_metadata,omitempty"`
	CallbackId      string           `json:"callback_id,omitempty"`
	ClearOnClose    *bool            `json:"clear_on_close,omitempty"`
	NotifyOnClose   *bool            `json:"notify_on_close,omitempty"`
	ExternalId      string           `json:"external_id,omitempty"`
	SubmitDisabled  *bool            `json:"submit_disabled,omitempty"`
}

func (data PrivateMetadata) MarshalJSON() ([]byte, error) {
	// encrypt
	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		return nil, err
	}

	// base64 encoding
	return json.Marshal(base64.StdEncoding.EncodeToString(encrypted))
}

type genericResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (api *API) PublishHomeView(user *User, blocks []block.Block) error {
	return api.PublishView(user, &View{
		Type: "home",
		Title: block.PlainText{
			Text:  "Hello, world!",
			Emoji: false,
		},
		Blocks: blocks,
	})
}

type publishViewRequest struct {
	UserId string `json:"user_id"`
	View   *View  `json:"view"`
}

func (api *API) PublishView(user *User, view *View) error {
	req := publishViewRequest{
		UserId: user.ID,
		View:   view,
	}

	// do request
	resp, err := api.doHTTPPostJSON("api/views.publish", nil, req)

	if err != nil {
		return fmt.Errorf("failed to send request to publish view: %v", err)
	}

	defer resp.Body.Close()

	// check result
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish view: status = %s", resp.Status)
	}

	// read response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to response body: %v", err)
	}

	// unmarshal generic response
	var gresp genericResponse

	err = json.Unmarshal(b, &gresp)
	if err != nil {
		// never reached
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if !gresp.OK {
		return fmt.Errorf("failed to publish view: %s", gresp.Error)
	}

	return nil
}

type openViewRequest struct {
	TriggerId string `json:"trigger_id"`
	View      *View  `json:"view"`
}

type openViewResponse struct {
	View map[string]interface{} `json:"view"`
	genericResponse
}

func (api *API) OpenView(triggerId string, view *View) (viewId string, err error) {
	req := openViewRequest{
		TriggerId: triggerId,
		View:      view,
	}

	// do request
	resp, err := api.doHTTPPostJSON("api/views.open", nil, req)

	if err != nil {
		return "", fmt.Errorf("failed to send request to open view: %v", err)
	}

	defer resp.Body.Close()

	// check result
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to open view: status = %s", resp.Status)
	}

	// read response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to response body: %v", err)
	}

	// unmarshal generic response
	var vResp openViewResponse

	err = json.Unmarshal(b, &vResp)
	if err != nil {
		// never reached
		return "", fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if !vResp.OK {
		return "", fmt.Errorf("failed to open view: %s", vResp.Error)
	}

	// extract id
	iId, ok := vResp.View["id"]
	if !ok {
		return "", fmt.Errorf("empty view id")
	}

	viewId, ok = iId.(string)
	if !ok {
		// never reached
		return "", fmt.Errorf("invalid viewId type: %v", reflect.TypeOf(iId))
	}

	return
}
