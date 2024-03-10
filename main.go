package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func main() {
	os.Setenv("TZ", "Europe/Paris")
	// CLI arguments
	listenAddr := flag.String("listen", "127.0.0.1:8080", "listen address")
	logLevel := flag.String("log-level", "info", "min level of logs to print")

	db_host := flag.String("db_host", "127.0.0.1", "Database host")
	db_port := flag.Int("db_port", 5432, "Database port")
	db_user := flag.String("db_user", "tfury-api", "Database user")
	db_password := flag.String("db_password", "change_me", "Database password")
	db_dbname := flag.String("db_name", "tfury.com", "Database name")
	db_type := flag.String("db_type", "postgres", "Database name")
	flag.Parse()

	// Connecting to database
	db_constring := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", *db_host, *db_port, *db_user, *db_password, *db_dbname)
	fmt.Println(db_constring)
	db, err := sql.Open(*db_type, db_constring)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := &AppCtx{
		listenAddr: *listenAddr,
		logLevel:   *logLevel,
		db:         db,
	}

	app, err := NewApp(ctx)
	if err != nil {
		logrus.Panic(err)
	}

	app.Run()
}
