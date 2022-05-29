package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/healthchecks/http"
	"github.com/projectkeas/sdks-service/server"
)

func main() {
	app := server.New("testing")

	app.WithRequiredConfigMap("config-1").WithInMemoryConfiguration("default", map[string]string{
		"test": "value",
	}).WithEnvironmentVariableConfiguration("DOTNET_")

	app.WithLivenessHealthCheck(http.NewHttpHealthCheck("https://google.com"))
	app.WithReadinessHealthCheck(http.NewHttpHealthCheck("https://bing.com"))

	app.ConfigureHandlers(func(f *fiber.App, server *server.Server) {
		f.Get("/", func(c *fiber.Ctx) error {
			value := server.GetConfiguration().GetStringValueOrDefault("log.level", "not set")
			return c.SendString(fmt.Sprintf("Hello, World 👋! Log Level is: %s", value))
		})
	})

	app.Build().Run()
}
