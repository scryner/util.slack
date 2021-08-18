package server

import (
	"github.com/labstack/echo/v4"
)

func Http(method, endpoint string, handlerFunc echo.HandlerFunc) handler {
	return func() (string, string, echo.HandlerFunc, bool) {
		return method, endpoint, handlerFunc, false
	}
}
