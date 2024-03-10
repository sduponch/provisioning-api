package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Device struct {
	Id            uuid.UUID `json:"uuid"`
	Token         []byte    `json:"bearer_token"`
	Data          []byte    `json:"-"`
	RemoteAddress string    `json:"remoteIP"`
	Expire        time.Time `json:"expire_at"`
}
type DeviceTokens map[string]Device

var subscribe_stmt *sql.Stmt

func (app *App) Subscribe(remoteAddress string, data []byte) (Device, error) {
	var device Device
	var err error
	device.Id = uuid.New()
	device.Data = data
	device.RemoteAddress = remoteAddress
	if subscribe_stmt == nil {
		subscribe_stmt, err = app.db.Prepare(`INSERT INTO "Devices" ("id", "data", "lastRemoteAddress", "lastSeen") VALUES ($1, $2, $3, $4)`)
		if err != nil {
			panic(err)
		}
	}
	_, err = subscribe_stmt.Query(&device.Id, &device.Data, &device.RemoteAddress, time.Now())
	if err != nil {
		return device, err
	}
	token, err := app.generateToken()
	if err != nil {
		return device, err
	}
	device.Token = token
	device.Expire = time.Now().UTC()
	return device, err
}

func (app *App) Associated(deviceID string) (string, error) {
	return deviceID, errors.New("not yey implemented")
}

func (app *App) Activate(deviceID string) error {
	return errors.New("not tyet implemented")
}
