package main

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/sirupsen/logrus"

	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	"crypto/rand"
	b64 "encoding/base64"

	"strings"
	"errors"
)

type AppCtx struct {
	listenAddr   string
	logLevel     string
}

// App holds the application
type User struct {
	login string
	scopes []string
}
type Token map[string]User
type App struct {
	listenAddr string
	// API Engine
	render *render.Render
	router *httprouter.Router
	logger *logrus.Logger
	tokens Token
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
		listenAddr:   ctx.listenAddr,
		render:       render.New(),
		router:       httprouter.New(),
		logger:       logger,
		tokens:	      make(Token),
	}, nil
}

func customBefore(entry *logrus.Entry, req *http.Request, remoteAddr string) *logrus.Entry {
//	reqHeadersBytes, _ := json.Marshal(req.Header)
	return entry.WithFields(logrus.Fields{
		"request": req.RequestURI,
		"method":  req.Method,
		"remote":  remoteAddr,
		"xff": req.Header.Get("X-Forwarded-For"),
		"rip": req.Header.Get("X-Real-IP"),
		// "auth": req.Header.Get("Authorization"),
		// "headers": string(reqHeadersBytes),
	})
}

func (app *App) Run() {
	// Template handler
	app.router.GET("/ping", app.Ping)
	app.router.GET("/auth", app.Auth)
	app.router.GET("/me", app.Me)

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

func (app *App) Me(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user, err := app.Bearer(r)
	if err != nil {
		app.render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	app.render.JSON(w, http.StatusOK, map[string]string{"user": user.login})
}

func (app *App) Bearer(r *http.Request) (User, error) {
        bearerToken := r.Header.Get("Authorization")
        if bearerToken == "" {
                return User{}, errors.New("bearer token is missing")
        }
        reqToken := strings.Split(bearerToken, " ")[1]
	token, err := b64.StdEncoding.DecodeString(reqToken)
        if(err != nil || len(token) != 256) {
		return User{}, errors.New("malformed bearer token")
        }
        user, ok := app.tokens[string(token)]
        if !ok {
                return User{}, errors.New("invalid bearer token")
        }
	return user, nil
}

func (app *App) Auth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	login := r.URL.Query().Get("login")
	pass  := r.URL.Query().Get("password")
	if (login == "admin" && pass == "change_me") {
		token, err := app.generateToken()
		if err != nil {
			app.renderError(w, err)
			return
		}
		scopes := []string{"read.all", "write.all"}
		app.tokens[string(token)] = User{login: login, scopes: scopes}
		app.render.JSON(w, http.StatusOK, map[string]string{"bearer_token":  b64.StdEncoding.EncodeToString(token)})
		return
	}
	app.render.JSON(w, http.StatusNotAcceptable, map[string]string{"message": "invalid credentials"})
}

func (app *App) generateToken() ([]byte, error) {
    bytes := make([]byte, 256)
    if _, err := rand.Read(bytes); err != nil {
        return []byte{}, err
    }
    return bytes, nil
}
