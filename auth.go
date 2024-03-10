package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type User struct {
	expire time.Time
	login  string
	scopes []string
}
type Token map[string]User

var login_stmt *sql.Stmt

func (app *App) Login(user string, password string) ([]byte, error) {
	token, err := app.generateToken()
	if err != nil {
		return []byte{}, err
	}
	// @todo in database
	if login_stmt == nil {
		login_stmt, err = app.db.Prepare(`SELECT "scopes" FROM "Users" WHERE "login"=$1 and "password"=$2`)
		if err != nil {
			panic(err)
		}
	}
	var scopes []string
	row := login_stmt.QueryRow(user, password)
	if err := row.Scan(pq.Array(&scopes)); err != nil { // scan will release the connection
		if err == sql.ErrNoRows {
			return []byte{}, errors.New("invalid credential")
		}
		panic(err)
	}
	login := user
	app.tokens[string(token)] = User{login: login, scopes: scopes, expire: time.Now().Add(2 * time.Hour)}
	return token, nil
}

func (app *App) generateToken() ([]byte, error) {
	bytes := make([]byte, 256)
	if _, err := rand.Read(bytes); err != nil {
		return []byte{}, err
	}
	return bytes, nil
}
