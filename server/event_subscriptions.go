package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/labstack/echo/v4"
)

type Authorizations []Authorization

func (auths Authorizations) IsBot() bool {
	var isBot bool

	for _, auth := range auths {
		isBot = isBot || auth.IsBot
	}

	return isBot
}

type Authorization struct {
	EnterpriseId        string `json:"enterprise_id"`
	TeamId              string `json:"team_id"`
	UserId              string `json:"user_id"`
	IsBot               bool   `json:"is_bot"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install"`
}

type Event map[string]interface{}

func (ev Event) Type() (typ, subType string, err error) {
	// get type
	iTyp, ok := ev["type"]
	if !ok {
		err = fmt.Errorf("type field is missing")
		return
	}

	if typ, ok = iTyp.(string); !ok {
		err = fmt.Errorf("invalid type field: %v, but string is needed", reflect.TypeOf(iTyp))
		return
	}

	// get sub-type
	if iSubTyp, ok := ev["subtype"]; ok {
		subType, _ = iSubTyp.(string)
	}

	return
}

type EventCallback struct {
	TeamId         string         `json:"team_id"`
	ApiAppId       string         `json:"api_app_id"`
	Event          Event          `json:"event"`
	Authorizations Authorizations `json:"authorizations"`
	EventContext   string         `json:"event_context"`
	EventTime      time.Time      `json:"event_time"`
}

func (cb *EventCallback) UnmarshalJSON(b []byte) error {
	var _cb EventCallback
	var m map[string]interface{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	for k, v := range m {
		switch k {
		case "team_id":
			_cb.TeamId = safeToString(v)

		case "api_app_id":
			_cb.ApiAppId = safeToString(v)

		case "event":
			if props, ok := v.(map[string]interface{}); ok {
				_cb.Event = props
			} else {
				return fmt.Errorf("invalid event structure")
			}

		case "authorizations":
			if vi, ok := v.([]interface{}); ok {
				var authorizations Authorizations

				for _, i := range vi {
					if props, ok := i.(map[string]interface{}); ok {
						var auth Authorization

						if err := unmarshalFromMap(props, &auth); err != nil {
							return err
						}

						authorizations = append(authorizations, auth)
					} else {
						return fmt.Errorf("invalid authorization structure: %v is provided", reflect.TypeOf(i))
					}
				}

				_cb.Authorizations = authorizations
			} else {
				return fmt.Errorf("invalid authorizations structure: %v is provided", reflect.TypeOf(v))
			}

		case "event_context":
			_cb.EventContext = safeToString(v)

		case "event_time":
			_cb.EventTime = time.Unix(safeToInt64(v), 0)
		}
	}

	*cb = _cb
	return nil
}

func safeToInt64(v interface{}) int64 {
	switch i := v.(type) {
	case int:
		return int64(i)
	case int64:
		return i
	case float64:
		return int64(i)
	default:
		return 0
	}
}

type EventHandler interface {
	HandleEvent(ctx Context, cb *EventCallback) error
}

func EventSubscriptions(endpoint string, handler EventHandler) handler {
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

			switch typ {
			case "url_verification":
				challenge, _ := props["challenge"].(string)
				return ctx.String(http.StatusOK, challenge)

			case "event_callback":
				// unmarshal to callback
				var cb EventCallback

				if err := unmarshalFromMap(props, &cb); err != nil {
					ctx.Logger().Errorf("failed to unmarshal from event callback: %v", err)
					return ctx.JSON(http.StatusBadRequest, slackError{
						ResponseType: "ephemeral",
						Text:         "I can't unmarshal event callback payload",
					})
				}

				// handle it
				go func() {
					if err := handler.HandleEvent(ctx, &cb); err != nil {
						ctx.Logger().Errorf("failed to handle event: %v", err)
					}
				}()

				return ctx.NoContent(http.StatusOK)
			default:
				ctx.Logger().Errorf("unhandled event type '%s'", typ)
				return ctx.NoContent(http.StatusOK)
			}
		}
	}
}

func unmarshalFromMap(m map[string]interface{}, v interface{}) error {
	// marshal to json
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// unmarshal from json
	return json.Unmarshal(b, v)
}
