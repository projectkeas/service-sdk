package main

import (
	"github.com/projectkeas/sdks-service/healthchecks/http"
	"github.com/projectkeas/sdks-service/server"
)

func main() {
	app := server.New("testing")

	app.WithConfigMap("config-1")

	app.WithLivenessHealthCheck(http.NewHttpHealthCheck("https://google.com"))

	app.RunDevelopment()
}
