package server

import (
	"bytes"
	"encoding/json"
	"iot-stats/model"
	"iot-stats/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	apiHeader   = "Api-Key"
	apiKey      = "123456"
	expiration  = 24
	login       = "test"
	password    = "test"
	deviceCount = 2
)

var (
	devicesFromMongo = []model.DeviceDto{
		model.DeviceDto{
			DeviceNumber: "1234",
			RegisterDate: time.Now(),
			Errors:       []model.DeviceErrorDto{},
		},
		model.DeviceDto{
			DeviceNumber: "1234",
			RegisterDate: time.Now(),
			Errors:       []model.DeviceErrorDto{},
		},
	}
	devicesAnswer = model.Devices{
		Devices: &devicesFromMongo,
		Total:   deviceCount,
	}
	creds = model.Credentials{
		Login:    login,
		Password: password,
	}
	deviceError = model.DeviceErrorDto{
		DeviceNumber: "123",
		ErrorName:    "electricity",
	}
	deviceDto = PostDevice{
		DeviceNumber: "123",
	}
)

type FakeMongoService struct{}

func (m *FakeMongoService) Connect() error { return nil }
func (m *FakeMongoService) GetAllDevices(skip int, limit int) (*[]model.DeviceDto, error) {
	return &devicesFromMongo, nil
}
func (m *FakeMongoService) GetDevicesCount() (int, error) { return deviceCount, nil }
func (m *FakeMongoService) RegisterDevice(deviceNumber string,
	registerDate time.Time) error {
	return nil
}
func (m *FakeMongoService) RegisterError(de *model.DeviceErrorDto) error { return nil }
func (m *FakeMongoService) GetDeviceByNumber(deviceNumber string) (*model.Device, error) {
	return nil, nil
}
func (m *FakeMongoService) GetCookieExp(login string) (*time.Time, error) {
	expDuration := time.Duration(expiration) * time.Hour
	expires := time.Now().Local().Add(expDuration)
	return &expires, nil
}
func (m *FakeMongoService) SetCookieExp(login string, expireTime time.Time) error { return nil }
func (m *FakeMongoService) SetCreds(creds model.Credentials) error                { return nil }
func (m *FakeMongoService) GetCreds(login string) (*model.Credentials, error) {
	crds := &model.Credentials{
		Login:    login,
		Password: utils.GenerateHash(password),
	}
	return crds, nil
}

func fakeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func testCookies(login string) *http.Cookie {
	value := map[string]string{
		"name": login,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		expDuration := time.Duration(expiration) * time.Hour
		cookie := http.Cookie{
			Name:    "session",
			Value:   encoded,
			Path:    "/",
			Expires: time.Now().Local().Add(expDuration),
		}
		return &cookie
	}
	return nil
}

type ServerTestSuite struct {
	suite.Suite
	ms    *FakeMongoService
	api   *Api
	web   *Web
	login *Login
}

func (suite *ServerTestSuite) SetupTest() {
	suite.ms = &FakeMongoService{}
	suite.api = newApi(apiKey, suite.ms)
	suite.web = newWeb(expiration, suite.ms)
	suite.login = newLogin(expiration, suite.ms)
}

func (suite *ServerTestSuite) TestCheckSessionWeb() {
	testRouter := gin.Default()
	testRouter.Use(suite.web.checkSession)
	testRouter.GET("/", fakeHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	// Testing unauthorized request
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusUnauthorized, rw.Code)
	req, _ = http.NewRequest("GET", "/", nil)
	cookie := testCookies(login)
	req.Header.Add("Cookie", cookie.String())
	rw = httptest.NewRecorder()
	// Testing cookie processing
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)
}

func (suite *ServerTestSuite) TestGetDevices() {
	testRouter := gin.Default()
	testRouter.GET("/:skip/:limit", suite.web.getDevices)
	req, _ := http.NewRequest("GET", "/0/10", nil)
	rw := httptest.NewRecorder()
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)
	jsonAnswer, _ := json.Marshal(devicesAnswer)
	assert.Equal(suite.T(), jsonAnswer, rw.Body.Bytes())
}

func (suite *ServerTestSuite) TestLogin() {
	testRouter := gin.Default()
	testRouter.POST("/", suite.login.loginHandler)
	crdsToPost, _ := json.Marshal(creds)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(crdsToPost))
	rw := httptest.NewRecorder()
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)
	respCookie := rw.Result().Cookies()
	cookieValue := make(map[string]string)
	err := cookieHandler.Decode("session", respCookie[0].Value, &cookieValue)
	assert.Equal(suite.T(), nil, err)
	cookieLogin := cookieValue["login"]
	assert.Equal(suite.T(), login, cookieLogin)
}

func (suite *ServerTestSuite) TestApi() {
	testRouter := gin.Default()
	testRouter.POST("/error", suite.api.errorReport)
	errToPost, _ := json.Marshal(deviceError)
	// Test error report
	req, _ := http.NewRequest("POST", "/error", bytes.NewReader(errToPost))
	rw := httptest.NewRecorder()
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)

	// Test register device
	testRouter.POST("/register", suite.api.registerDevice)
	deviceToPost, _ := json.Marshal(deviceDto)
	req, _ = http.NewRequest("POST", "/register", bytes.NewReader(deviceToPost))
	rw = httptest.NewRecorder()
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)
}

func (suite *ServerTestSuite) TestCheckApiKey() {
	testRouter := gin.Default()
	testRouter.Use(suite.api.checkApiKey)
	testRouter.GET("/", fakeHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	// Testing unauthorized request
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusUnauthorized, rw.Code)
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add(apiHeader, apiKey)
	rw = httptest.NewRecorder()
	// Testing apikey processing
	testRouter.ServeHTTP(rw, req)
	assert.Equal(suite.T(), http.StatusOK, rw.Code)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
