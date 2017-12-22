package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Device struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	DeviceNumber string        `bson:"device_number"`
	RegisterDate time.Time     `bson:"register_date"`
}

type DeviceDto struct {
	DeviceNumber string           `bson:"device_number" json:"device-number"`
	RegisterDate time.Time        `bson:"register_date" json:"register-date"`
	Errors       []DeviceErrorDto `bson:"errors" json:"errors"`
}

type DeviceErrorDto struct {
	ErrorName    string `bson:"error_name" json:"error-name"`
	DeviceNumber string `bson:"device_number" json:"device-number"`
}

type DeviceError struct {
	ErrorName    string        `bson:"error_name"`
	DeviceNumber string        `bson:"device_number"`
	Date         time.Time     `bson:"date"`
	DeviceId     bson.ObjectId `bson:"device_id"`
}

type Devices struct {
	Devices *[]DeviceDto `json:"devices"`
	Total   int          `json:"total"`
}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Session struct {
	ID     string    `json:"id"`
	Expire time.Time `json:"expire"`
}

type Cookie struct {
	Login  string    `bson:"login"`
	Expire time.Time `bson:"expire"`
}
