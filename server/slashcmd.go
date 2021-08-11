package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/scryner/util.slack/block"
)

type SlashCommandRequest struct {
	ChannelId   string
	ChannelName string
	Command     string
	ResponseUrl string
	TeamDomain  string
	TeamId      string
	Text        string
	Token       string
	TriggerId   string
	UserId      string
	UserName    string
}

type SlashCommandHandler interface {
	HandleCommand(Context, *SlashCommandRequest) (block.Message, error)
}

func SlashCommand(endpoint string, cmdHandler SlashCommandHandler) handler {
	return func() (path string, handlerFunc echo.HandlerFunc) {
		return endpoint, func(ctx echo.Context) error {
			// get request body
			_reqBody := ctx.Get("reqBody")
			reqBody, ok := _reqBody.([]byte)
			if !ok {
				// actually, never happened
				ctx.Logger().Errorf("empty request body")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your request body",
				})
			}

			// get command value
			formVals, err := url.ParseQuery(string(reqBody))
			if err != nil {
				ctx.Logger().Errorf("failed to read form param: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your form params",
				})
			}

			request := &SlashCommandRequest{
				ChannelId:   formVals.Get("channel_id"),
				ChannelName: formVals.Get("channel_name"),
				Command:     formVals.Get("command"),
				ResponseUrl: formVals.Get("response_url"),
				TeamDomain:  formVals.Get("team_domain"),
				TeamId:      formVals.Get("team_id"),
				Text:        formVals.Get("text"),
				Token:       formVals.Get("token"),
				TriggerId:   formVals.Get("trigger_id"),
				UserId:      formVals.Get("user_id"),
				UserName:    formVals.Get("user_name"),
			}

			msg, err := cmdHandler.HandleCommand(ctx, request)

			if err != nil {
				ctx.Logger().Errorf("failed to handle request: %v", err)
				return ctx.JSON(http.StatusOK, slackError{
					ResponseType: "ephemeral",
					Text:         fmt.Sprintf("I can't handle request due to: %v", err),
				})
			}

			return ctx.JSON(http.StatusOK, msg)
		}
	}
}
