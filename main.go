package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	os.Setenv("TZ", "Europe/Paris")
	// CLI arguments
	listenAddr := flag.String("listen", "127.0.0.1:8080", "listen address")
	logLevel := flag.String("log-level", "info", "min level of logs to print")
	flag.Parse()

	ctx := &AppCtx{
		listenAddr:   *listenAddr,
		logLevel:     *logLevel,
	}
	app, err := NewApp(ctx)
	if err != nil {
		logrus.Panic(err)
	}

	app.Run()
}
