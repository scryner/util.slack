package server

import (
	"github.com/labstack/echo/v4"
)

type Context interface {
	Logger() echo.Logger
	Get(key string) interface{}
	Set(key string, val interface{})
}
