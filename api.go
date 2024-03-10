package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/sirupsen/logrus"

	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

type AppCtx struct {
	listenAddr string
	logLevel   string
	db         *sql.DB
}

// App holds the application
type App struct {
	listenAddr string
	// API Engine
	render  *render.Render
	router  *httprouter.Router
	logger  *logrus.Logger
	tokens  UserTokens
	devices DeviceTokens
	db      *sql.DB
}

func NewApp(ctx *AppCtx) (*App, error) {
	// Get the logrus level
	lvl, err := logrus.ParseLevel(ctx.logLevel)
	if err != nil {
		return nil, err
	}

	// Setup the logger
	logger := logrus.New()
	logger.Level = lvl
	logger.Out = os.Stderr
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	return &App{
		listenAddr: ctx.listenAddr,
		render:     render.New(),
		router:     httprouter.New(),
		logger:     logger,
		tokens:     make(UserTokens),
		devices:    make(DeviceTokens),
		db:         ctx.db,
	}, nil
}

func customBefore(entry *logrus.Entry, req *http.Request, remoteAddr string) *logrus.Entry {
	//	reqHeadersBytes, _ := json.Marshal(req.Header)
	return entry.WithFields(logrus.Fields{
		"request": req.RequestURI,
		"method":  req.Method,
		"remote":  remoteAddr,
		"xff":     req.Header.Get("X-Forwarded-For"),
		"rip":     req.Header.Get("X-Real-IP"),
		// "auth": req.Header.Get("Authorization"),
		// "headers": string(reqHeadersBytes),
	})
}

func (app *App) Run() {
	// Template handler
	app.router.GET("/ping", app.Ping)
	app.router.GET("/auth", app.Auth)
	app.router.GET("/me", app.Me)

	//app.router.GET("/service/associated", app.Devices)
	//app.router.POST("/service/confirm", app.Devices)

	app.router.POST("/subscribe", app.SubscribeDevice)
	app.router.GET("/service/associated", app.AssociateService)
	app.router.POST("/service/confirm", app.ConfirmService)

	//app.router.GET("/devices/:uuid/:subpath", app.Devices)
	//app.router.PUT("/devices/:uuid/:subpath", app.Devices)

	//app.router.POST("/devices/:uuid/events", app.Devices)

	//app.router.PUT("/service/:uuid/:subpath", app.Devices)
	//app.router.POST("/service/:uuid/:subpath", app.Devices)

	//app.router.GET("/actions/todo", app.DeviceAction)         // endpoint where registred devices poll todo, return a todo with as associated id
	//app.router.POST("/actions/:id", app.DeviceActionUpdate)	// endpoint where registred devices inform the status of the current todo (doing, done, error)

	//app.router.GET("/locations", app.Me)
	//app.router.GET("/locations/:uuid", app.Me)

	//app.router.GET("/cabinets", app.Me)
	//app.router.GET("/cabinets/:uuid", app.Me)

	app.router.NotFound = http.FileServer(http.Dir("data"))
	// Negroni handler
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	middle := negronilogrus.NewCustomMiddleware(logrus.InfoLevel, app.logger.Formatter, "httpServer")
	middle.Before = customBefore
	n.Use(middle)
	n.UseHandler(app.router)

	app.logger.Infof("listening on http://%s", app.listenAddr)

	n.Run(app.listenAddr)
}

// renderError is an helper to render a server error
func (app *App) renderError(w http.ResponseWriter, err error) {
	app.logger.Error(err)
	app.render.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

// renderNotFound renders a 404 response
func (app *App) renderNotFound(w http.ResponseWriter, err error) {
	app.logger.Warn(err)
	app.render.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
}

// renderError is an helper to render a server error
func (app *App) returnError(w http.ResponseWriter, err error) {
	app.render.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
}

// Ping respond to /ping
func (app *App) Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	app.render.JSON(w, http.StatusOK, map[string]string{"message": "pong"})
}
