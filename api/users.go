package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type User struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	TzOffset int    `json:"tz_offset"`

	Profile struct {
		Email string `json:"email"`
	} `json:"profile"`
	DMChannel string `json:"-"`
}

type userInfoResponse struct {
	OK   bool `json:"ok"`
	User User `json:"user"`
}

func (api *API) SearchUserByEmail(email string) (*User, error) {
	// try to get from cache
	iUser, ok, err := api.emailToUserCache.Get(email)
	if err != nil {
		return nil, err
	}

	if ok {
		user := iUser.(*User)
		return user, nil
	}

	// fallback
	params := make(url.Values)
	params.Set("email", email)

	// request
	resp, err := api.doHTTPGet("api/users.lookupByEmail", params)
	if err != nil {
		return nil, fmt.Errorf("failed to do search user: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to do search user: status code = %s", resp.Status)
	}

	// parse user info
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search user body: %v", err)
	}

	var lookupResp userInfoResponse
	err = json.Unmarshal(body, &lookupResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %v", err)
	}

	user := lookupResp.User
	if user.ID == "" {
		return nil, fmt.Errorf("no matching email")
	}

	api.emailToUserCache.Set(email, &user)
	api.idToUserCache.Set(user.ID, &user)

	return &user, nil
}

func (api *API) GetUserInfo(id string) (*User, error) {
	// try to get from cache
	iUser, ok, err := api.idToUserCache.Get(id)
	if err != nil {
		return nil, err
	}

	if ok {
		user := iUser.(*User)
		return user, nil
	}

	// fallback
	params := make(url.Values)
	params.Set("user", id)

	// request
	resp, err := api.doHTTPGet("api/users.info", params)
	if err != nil {
		return nil, fmt.Errorf("failed to do get user info: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to do get user info: status code = %s", resp.Status)
	}

	// parse user info
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info body: %v", err)
	}

	var userInfoResp userInfoResponse
	err = json.Unmarshal(body, &userInfoResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %v", err)
	}

	user := userInfoResp.User
	if user.ID == "" || user.Profile.Email == "" {
		return nil, fmt.Errorf("no matching user")
	}

	api.idToUserCache.Set(id, &user)
	api.emailToUserCache.Set(user.Profile.Email, &user)

	return &user, nil
}
