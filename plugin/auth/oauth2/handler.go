package oauth2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

var errInvalidOauthState = errors.New("invalid oauth state")

const (
	sessionDuration = 24 * time.Hour
	stateDuration   = 5 * time.Minute
)

type oauth2Handlers struct {
	pluginCfg *oAuth2Cfg
	oauth2    *oauth2.Config
	kvStore   core.StoreKV
}

func (oauth2Handlers oauth2Handlers) logout(reqCtx echo.Context) error {
	setUserSessionCookie(reqCtx, "logged-out", 1, time.Unix(0, 0))
	return reqCtx.Redirect(http.StatusSeeOther, "/")
}

func (oauth2Handlers oauth2Handlers) loggedOut(reqCtx echo.Context) error {
	return reqCtx.HTML(http.StatusOK, loggedOutHTML)
}

func (oauth2Handlers oauth2Handlers) login(reqCtx echo.Context) error {
	csrf := core.GetCSRFToken(reqCtx.Request().Context())

	err := oauth2Handlers.kvStore.Set(
		reqCtx.Request().Context(),
		"oauth2-plugin",
		fmt.Sprintf("state:%s", csrf),
		true,
		stateDuration,
	)
	if err != nil {
		return err
	}

	url := oauth2Handlers.oauth2.AuthCodeURL(csrf, oauth2.AccessTypeOnline)

	return reqCtx.Redirect(http.StatusSeeOther, url)
}

func (oauth2Handlers oauth2Handlers) callback(reqCtx echo.Context) error {
	code := reqCtx.FormValue("code")
	state := reqCtx.FormValue("state")
	ctx := reqCtx.Request().Context()

	var stateExist *bool

	err := oauth2Handlers.kvStore.Get(ctx, "oauth2-plugin", fmt.Sprintf("state:%s", state), &stateExist)
	if err != nil {
		return err
	}

	if stateExist == nil {
		return errInvalidOauthState
	}

	token, err := oauth2Handlers.oauth2.Exchange(ctx, code)
	if err != nil {
		return err
	}

	client := oauth2Handlers.oauth2.Client(ctx, token)

	email, err := getUserEmail(ctx, client, oauth2Handlers.pluginCfg.UserInfo)
	if err != nil {
		return err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": email,
		"exp": jwt.NewNumericDate(time.Now().Add(sessionDuration)),
	})

	encodedToken, err := jwtToken.SignedString([]byte(oauth2Handlers.pluginCfg.JWTSecret))
	if err != nil {
		return err
	}

	setUserSessionCookie(reqCtx, encodedToken, int(sessionDuration.Seconds()), time.Now().Add(sessionDuration))

	return reqCtx.Redirect(http.StatusFound, "/")
}

func setUserSessionCookie(c echo.Context, value string, maxAge int, expires time.Time) {
	c.SetCookie(&http.Cookie{
		Name:     "user-session",
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expires,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: true,
	})
}

func getUserEmail(ctx context.Context, client *http.Client, userInfoCfg userInfoCfg) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoCfg.URL, http.NoBody)
	if err != nil {
		slog.Error("err", slog.Any("err", err))
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("err", slog.Any("err", err))
		return "", err
	}
	defer resp.Body.Close()

	var userInfoBody map[string]interface{}

	bodyReader := getBodyDecoder(resp, userInfoCfg.BodyDecoder)
	if decodeErr := json.NewDecoder(bodyReader).Decode(&userInfoBody); decodeErr != nil {
		return "", decodeErr
	}

	if verifyErr := verifyEmail(userInfoCfg, userInfoBody); verifyErr != nil {
		return "", verifyErr
	}

	email, jsonErr := getByJSONPath[string](userInfoBody, strings.Split(userInfoCfg.EmailPath, "."))
	if jsonErr != nil {
		return "", jsonErr
	}

	return email, nil
}
func getBodyDecoder(resp *http.Response, bodyDecoder string) io.Reader {
	switch bodyDecoder {
	case "base64":
		return base64.NewDecoder(base64.URLEncoding, resp.Body)
	default:
		return resp.Body
	}
}

var (
	errEmptyPath            = errors.New("empty path")
	errValueNotFound        = errors.New("value not found")
	errInvalidTypeOrMissing = errors.New("key of invalid type or missing")
	errEmailNotVerified     = errors.New("email not verified")
)

func verifyEmail(userInfoCfg userInfoCfg, userInfoBody map[string]interface{}) error {
	if userInfoCfg.EmailVerifiedPath != "" {
		emailVerified, err := getByJSONPath[any](userInfoBody, strings.Split(userInfoCfg.EmailVerifiedPath, "."))
		if err != nil {
			return fmt.Errorf("failed to retrieve email verification status: %w", err)
		}

		// Check for boolean or string "true"
		verified, isBool := emailVerified.(bool)
		verifiedStr, isString := emailVerified.(string)

		if (!isBool || !verified) && (!isString || verifiedStr != "true") {
			return errEmailNotVerified
		}
	}

	return nil
}

func getByJSONPath[T any](jsonData map[string]interface{}, path []string) (T, error) {
	var empty T
	if len(path) == 0 {
		return empty, errEmptyPath
	}

	val, ok := jsonData[path[0]]
	if !ok {
		return empty, errValueNotFound
	}

	// Recurse into nested maps if needed
	if subMap, isSubMap := val.(map[string]interface{}); isSubMap {
		return getByJSONPath[T](subMap, path[1:])
	}

	// Check if the value is of the expected type
	if result, isExpectedType := val.(T); isExpectedType {
		return result, nil
	}

	return empty, errInvalidTypeOrMissing
}
