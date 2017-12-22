package service

import (
	"fmt"
	"iot-stats/model"
	"net"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Config struct {
	Port     string
	Host     string
	User     string
	Password string
	Database string
}

func (c *Config) GetAddr() string {
	address := net.JoinHostPort(c.Host, c.Port)
	url := fmt.Sprintf("%s:%s@%s/%s", c.User, c.Password, address, c.Database)
	return url
}

type MongoInterface interface {
	Connect() error
	GetAllDevices(skip int, limit int) (*[]model.DeviceDto, error)
	GetDevicesCount() (int, error)
	RegisterDevice(deviceNumber string,
		registerDate time.Time) error
	RegisterError(de *model.DeviceErrorDto) error
	GetDeviceByNumber(deviceNumber string) (*model.Device, error)
	GetCookieExp(login string) (*time.Time, error)
	SetCookieExp(login string, expireTime time.Time) error
	SetCreds(creds model.Credentials) error
	GetCreds(login string) (*model.Credentials, error)
}

type MongoService struct {
	cfg *Config
	db  *mgo.Database
}

const (
	deviceCollection = "devices"
	cookieCollection = "cookies"
	userCollection   = "users"
	errorCollection  = "errors"
)

func NewMongoService(cfg *Config) *MongoService {
	m := &MongoService{cfg: cfg}
	return m
}

// Connect establish a connection to database
func (m *MongoService) Connect() error {
	session, err := mgo.Dial(m.cfg.GetAddr())
	if err != nil {
		return err
	}
	m.db = session.DB(m.cfg.Database)
	return nil
}

// GetAllDevices find list of devices
func (m *MongoService) GetAllDevices(skip int, limit int) (*[]model.DeviceDto,
	error) {
	deviceStore := m.db.C(deviceCollection)
	info := make([]model.DeviceDto, limit, limit)
	err := deviceStore.Pipe([]bson.M{
		bson.M{"$lookup": bson.M{"from": errorCollection, "localField": "_id",
			"foreignField": "device_id", "as": errorCollection}},
		bson.M{"$limit": limit},
		bson.M{"$skip": skip},
	}).All(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (m *MongoService) GetDevicesCount() (int, error) {
	deviceStore := m.db.C(deviceCollection)
	n, err := deviceStore.Find(bson.M{}).Count()
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (m *MongoService) RegisterDevice(deviceNumber string,
	registerDate time.Time) error {
	deviceStore := m.db.C(deviceCollection)
	colQuerier := bson.M{"device_number": deviceNumber}
	change := bson.M{"$set": bson.M{"device_number": deviceNumber, "register_date": registerDate}}
	if _, err := deviceStore.Upsert(colQuerier, change); err != nil {
		return err
	}
	return nil
}

func (m *MongoService) RegisterError(de *model.DeviceErrorDto) error {
	device, err := m.GetDeviceByNumber(de.DeviceNumber)
	if err != nil {
		return err
	}
	deviceError := &model.DeviceError{
		ErrorName:    de.ErrorName,
		DeviceNumber: de.DeviceNumber,
		Date:         time.Now(),
		DeviceId:     device.ID,
	}
	errorStore := m.db.C(errorCollection)
	if err := errorStore.Insert(deviceError); err != nil {
		return err
	}
	return nil
}

func (m *MongoService) GetDeviceByNumber(deviceNumber string) (*model.Device, error) {
	sessionStore := m.db.C(deviceCollection)
	device := &model.Device{}
	err := sessionStore.Find(bson.M{"device_number": deviceNumber}).One(device)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (m *MongoService) GetCookieExp(login string) (*time.Time, error) {
	sessionStore := m.db.C(cookieCollection)
	cookie := model.Cookie{}
	if err := sessionStore.Find(bson.M{"login": login}).One(&cookie); err != nil {
		return nil, err
	}
	return &cookie.Expire, nil
}

func (m *MongoService) SetCookieExp(login string, expireTime time.Time) error {
	sessionStore := m.db.C(cookieCollection)
	colQuerier := bson.M{"login": login}
	change := bson.M{"$set": bson.M{"expire": expireTime}}
	if _, err := sessionStore.Upsert(colQuerier, change); err != nil {
		return err
	}
	return nil
}

func (m *MongoService) SetCreds(creds model.Credentials) error {
	usersStore := m.db.C(userCollection)
	colQuerier := bson.M{"login": creds.Login}
	change := bson.M{"$set": bson.M{"password": creds.Password}}
	if _, err := usersStore.Upsert(colQuerier, change); err != nil {
		return err
	}
	return nil
}

func (m *MongoService) GetCreds(login string) (*model.Credentials, error) {
	usersStore := m.db.C(userCollection)
	creds := model.Credentials{}
	if err := usersStore.Find(bson.M{"login": login}).One(&creds); err != nil {
		return nil, err
	}
	return &creds, nil
}
