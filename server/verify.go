package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/labstack/echo/v4"
)

// Verifier verifies whether request was came from Slack
type Verifier struct {
	signingSecret   string
	verifyTimestamp func(int64) bool
}

// Verify do verifying request
func (v *Verifier) Verify(timestamp int64, reqSignature, reqBody string) error {
	// verify timestamp
	var ok bool

	if v.verifyTimestamp == nil {
		ok = defaultVerifyTimestamp(timestamp)
	} else {
		ok = v.verifyTimestamp(timestamp)
	}

	if !ok {
		return errors.New("failed to verify timestamp")
	}

	// make sig_basestring
	sigBaseString := fmt.Sprintf("v0:%d:%s", timestamp, reqBody)

	// calculate my signature
	hm := hmac.New(sha256.New, []byte(v.signingSecret))
	hm.Write([]byte(sigBaseString))

	mySignature := fmt.Sprintf("v0=%x", hm.Sum(nil))

	// compare signature
	ok = hmac.Equal([]byte(mySignature), []byte(reqSignature))
	if !ok {
		return errors.New("failed to compare signatures")
	}

	return nil
}

func (v *Verifier) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// verify token
			reqTimestamp := fromHeaderAsInt64(ctx.Request().Header, "X-Slack-Request-Timestamp")
			reqSignature := ctx.Request().Header.Get("X-Slack-Signature")

			reqBody, err := ioutil.ReadAll(ctx.Request().Body)
			if err != nil {
				ctx.Logger().Errorf("failed to read request body: %v", err)
				return echo.ErrBadRequest
			}

			ctx.Set("reqBody", reqBody)

			err = v.Verify(reqTimestamp, reqSignature, string(reqBody))
			if err != nil {
				ctx.Logger().Errorf("failed to verify request: %v", err)
				return echo.ErrForbidden
			}

			// verified
			return next(ctx)
		}
	}
}

// NewVerifier to create verifier
func NewVerifier(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret:   signingSecret,
		verifyTimestamp: defaultVerifyTimestamp,
	}
}

func defaultVerifyTimestamp(timestamp int64) bool {
	ti := time.Unix(timestamp, 0)
	diff := int64(time.Now().Sub(ti))

	if diff < 0 {
		diff *= -1
	}

	if diff > int64(time.Minute*5) {
		return false
	}

	return true
}
