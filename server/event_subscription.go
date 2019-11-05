package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"github.com/scryner/util.slack/msgfmt"
)

type EventHandler interface {
	HandleEvent(props map[string]interface{}) error
}

type eventHandlerDef func() (string, EventHandler)

func EventHandlerDef(event string, handler EventHandler) eventHandlerDef {
	return func() (string, EventHandler) {
		return event, handler
	}
}

func EventSubscriptions(endpoint string, handlerDefs ...eventHandlerDef) handler {
	// registering handler
	handlers := make(map[string]EventHandler)

	for _, handlerDef := range handlerDefs {
		event, handler := handlerDef()
		handlers[event] = handler
	}

	return func() (path string, handlerFunc echo.HandlerFunc) {
		return endpoint, func(ctx echo.Context) error {
			// get request body
			reqBody, ok := ctx.Get("reqBody").([]byte)
			if !ok {
				// actually, never happened
				ctx.Logger().Errorf("empty request body")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your request body",
				})
			}

			// unmarshal to map
			var props map[string]interface{}

			err := json.Unmarshal(reqBody, &props)
			if err != nil {
				ctx.Logger().Errorf("failed to unmarshal request to json: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't understand your request body",
				})
			}

			// get event type
			typ, ok := props["type"].(string)
			if !ok || typ == "" {
				ctx.Logger().Errorf("event type was nil")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't determine event type",
				})
			}

			if typ == "url_verification" {
				challenge, _ := props["challenge"].(string)
				return ctx.String(http.StatusOK, challenge)
			}

			// dispatch event
			h, ok := handlers[typ]
			if !ok {
				ctx.Logger().Errorf("can't dispatch event '%s': not registered event", typ)
				return ctx.JSON(http.StatusOK, msgfmt.PlainText{
					Text:         fmt.Sprintf("I don't know what I do when event '%s'", typ),
				})
			}

			// handle it
			go func() {
				if err = h.HandleEvent(props); err != nil {
					ctx.Logger().Errorf("failed to handle request: %v", err)
				}
			}()

			return ctx.NoContent(http.StatusOK)
		}
	}
}
