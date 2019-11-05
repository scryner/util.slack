package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/scryner/util.slack/msgfmt"
)

type openDMChannelResponse struct {
	OK      bool `json:"ok"`
	Channel struct {
		ID string `json:"id"`
	} `json:"channel"`
	Error string `json:"error"`
}

func (api *API) openDMChannel(user *User) (string, error) {
	if user.DMChannel != "" {
		return user.DMChannel, nil
	}

	// fallback
	params := make(url.Values)
	params.Set("users", user.ID)

	resp, err := api.doHTTPPost("api/conversations.open", params)
	if err != nil {
		return "", fmt.Errorf("failed to request open dm channel: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to request open dm channel: status = %s", resp.Status)
	}

	// try to parse result
	var openDMChannelResp openDMChannelResponse

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(b, &openDMChannelResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if !openDMChannelResp.OK {
		return "", fmt.Errorf("failed to open dm channel: %s", openDMChannelResp.Error)
	}

	channelID := openDMChannelResp.Channel.ID

	if channelID == "" {
		return "", fmt.Errorf("empty channel id")
	}

	// set to cache
	user.DMChannel = channelID
	api.emailToUserCache.Set(user.Profile.Email, user)
	api.idToUserCache.Set(user.ID, user)

	return channelID, nil
}

type ChatMessage struct {
	ChannelID        string         `json:"channel"`
	NotificationText string         `json:"text"`
	Blocks           []msgfmt.Block `json:"blocks"`
}

type postMessageResponse struct {
	OK        bool   `json:"ok"`
	ChannelId string `json:"channel"`
	Timestamp string `json:"ts"`
}

func (api *API) PostMessage(email string, msg *ChatMessage) (string, string, error) {
	// find user
	user, err := api.SearchUserByEmail(email)
	if err != nil {
		return "", "", fmt.Errorf("failed to search user '%s': %v", email, err)
	}

	// open DM channel
	dmChannel, err := api.openDMChannel(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to open DM channel for '%s': %v", user.Profile.Email, err)
	}

	msg.ChannelID = dmChannel

	// post message
	resp, err := api.doHTTPPostJSON("api/chat.postMessage", nil, msg)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request to post message: %v", err)
	}

	defer resp.Body.Close()

	// check result
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to post message: status = %s", resp.Status)
	}

	// read response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to response body: %v", err)
	}

	// unmarshal response
	var postMsgResp postMessageResponse

	err = json.Unmarshal(b, &postMsgResp)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if !postMsgResp.OK {
		return "", "", fmt.Errorf("failed to post message: resp = %s", string(b))
	}

	// return result
	return postMsgResp.ChannelId, postMsgResp.Timestamp, nil
}

type deleteMessageRequest struct {
	ChannelID string `json:"channel"`
	Timestamp string `json:"ts"`
	AsUser    bool   `json:"as_user"`
}

func (api *API) DeleteMessage(channelId, timestamp string, asUser bool) error {
	resp, err := api.doHTTPPostJSON("api/chat.delete", nil, &deleteMessageRequest{
		ChannelID: channelId,
		Timestamp: timestamp,
		AsUser:    asUser,
	})
	if err != nil {
		return fmt.Errorf("failed to send request to delete message: %v", err)
	}

	defer resp.Body.Close()

	// check result
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete message: status = %s", resp.Status)
	}

	return nil
}
