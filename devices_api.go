package main

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"encoding/json"
	"errors"
)

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

// endpoint where unconfigured devices subscribes, sending their hardware info like model, mac address, firmware version...
// and return a device uuid, permanent token, the associated service uuid if any and a bearer token
func (app *App) SubscribeDevice(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		app.renderError(w, err)
		return
	}
	dev, err := app.Subscribe(ReadUserIP(r), b)
	app.logger.Info(dev)
	if err != nil {
		app.renderError(w, err)
	}
	json, err := json.Marshal(dev)
	app.logger.Info(string(json), err)
	app.render.JSON(w, http.StatusInternalServerError, dev)
}

// return a service uuid, retry later if no service is provided
func (app *App) AssociateService(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	app.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": errors.New("AssociateService").Error()})
}

// activate service
func (app *App) ConfirmService(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	app.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": errors.New("ConfirmService").Error()})
}
