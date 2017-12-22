package server

import (
	"encoding/json"
	"iot-stats/model"
	"iot-stats/service"
	"iot-stats/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

type Login struct {
	ms  service.MongoInterface
	exp int
}

func newLogin(exp int, ms service.MongoInterface) *Login {
	return &Login{exp: exp, ms: ms}
}

func (l *Login) loginHandler(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	var crds model.Credentials
	err := decoder.Decode(&crds)
	if err != nil {
		internalError(c, "marshalling error",
			"marshalling error "+err.Error())
	}
	password := crds.Password
	login := crds.Login
	creds, err := l.ms.GetCreds(login)
	if err != nil {
		internalError(c, "database error",
			"database error "+err.Error())
	}
	if creds.Password == utils.GenerateHash(password) && creds.Login == login {
		l.setSession(login, c.Writer)
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	} else {
		c.Status(http.StatusUnauthorized)
	}
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func (l *Login) setSession(login string, response http.ResponseWriter) {
	value := map[string]string{
		"login": login,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		expDuration := time.Duration(l.exp) * time.Hour
		cookie := &http.Cookie{
			Name:    "session",
			Value:   encoded,
			Path:    "/",
			Expires: time.Now().Local().Add(expDuration),
		}
		l.ms.SetCookieExp(login, cookie.Expires)
		http.SetCookie(response, cookie)
	}
}
