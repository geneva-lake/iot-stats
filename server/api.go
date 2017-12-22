package server

import (
	"encoding/json"
	"iot-stats/model"
	"iot-stats/service"
	"iot-stats/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const ApiKey = "Api-Key"

type Api struct {
	apiKey string
	ms     service.MongoInterface
}

func newApi(apiKey string, ms service.MongoInterface) *Api {
	return &Api{apiKey: apiKey, ms: ms}
}

type PostDevice struct {
	DeviceNumber string `json:"device-number"`
}

// Checking api key in request
func (a *Api) checkApiKey(c *gin.Context) {
	ak := c.Request.Header.Get(ApiKey)
	if ak == "" {
		pleaseAuth(c, "No api key")
	} else if ak != a.apiKey {
		pleaseAuth(c, "Wrong api key")
	}
	c.Writer.Header().Add("Content-Type", "application/json")
	c.Next()
}

// Report about error in iot device
func (a *Api) errorReport(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	de := &model.DeviceErrorDto{}
	err := decoder.Decode(&de)
	if err != nil {
		internalError(c, "marshalling error",
			"marshalling error "+err.Error())
	}
	if err = a.ms.RegisterError(de); err != nil {
		internalError(c, "database error",
			"database error "+err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Error registered"})
}

// Registering device at server
func (a *Api) registerDevice(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	var postDevice PostDevice
	err := decoder.Decode(&postDevice)
	if err != nil {
		internalError(c, "marshalling error",
			"marshalling error"+err.Error())
	}
	if err = a.ms.RegisterDevice(postDevice.DeviceNumber, time.Now()); err != nil {
		internalError(c, "register err",
			"register err"+err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Device registered"})
}

func pleaseAuth(c *gin.Context, msg string) {
	if msg != "" {
		utils.Log().Infoln(msg)
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Access denied"})
	c.Abort()
}
