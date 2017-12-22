package server

import (
	"encoding/json"
	"iot-stats/model"
	"iot-stats/service"
	"iot-stats/utils"
	"net/http"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var authenticated bool = false

type Web struct {
	ms  service.MongoInterface
	exp int
}

func newWeb(exp int, ms service.MongoInterface) *Web {
	return &Web{ms: ms, exp: exp}
}

func (w *Web) checkSession(c *gin.Context) {
	cookieValue := make(map[string]string)
	if cookie, err := c.Request.Cookie("session"); err == nil {
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			login := cookieValue["login"]
			expire, err := w.ms.GetCookieExp(login)
			if err != nil {
				internalError(c, "databse error",
					"mongo web err "+err.Error())
			}
			if (*expire).Sub(time.Now().Local()) < 0 {
				pleaseAuth(c, "")
			}
			w.setSession(login, c.Writer)
			c.Writer.Header().Add("Content-Type", "application/json")
			c.Next()
		} else {
			pleaseAuth(c, "cookieSession err"+err.Error())
		}
	} else {
		pleaseAuth(c, "cookieSession err "+err.Error())
	}
}

func (w *Web) setSession(login string, response http.ResponseWriter) {
	value := map[string]string{
		"name": login,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		expDuration := time.Duration(w.exp) * time.Hour
		cookie := &http.Cookie{
			Name:    "session",
			Value:   encoded,
			Path:    "/",
			Expires: time.Now().Local().Add(expDuration),
		}
		w.ms.SetCookieExp(login, cookie.Expires)
		http.SetCookie(response, cookie)
	}
}

//Get list of registered devices
func (w *Web) getDevices(c *gin.Context) {
	skip, err := strconv.Atoi(c.Param("skip"))
	limit, err := strconv.Atoi(c.Param("limit"))
	if err != nil {
		utils.Log().Infoln("web err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
	}
	total, err := w.ms.GetDevicesCount()
	if err != nil {
		internalError(c, "databse error", "web err "+err.Error())
	}
	devices, err := w.ms.GetAllDevices(skip, limit)
	jDev := model.Devices{}
	if err != nil {
		internalError(c, "databse error", "web err "+err.Error())
	}
	jDev.Devices = devices
	jDev.Total = total
	jsonM, err := json.Marshal(&jDev)
	if err != nil {
		internalError(c, "marshaling error",
			"marshaling error "+err.Error())
	}
	c.String(http.StatusOK, string(jsonM))
}

func internalError(c *gin.Context, msgToSend, msgToLog string) {
	utils.Log().Infoln(msgToLog)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msgToSend})
	c.Abort()
}
