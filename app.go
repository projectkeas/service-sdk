package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks/http"
	"github.com/projectkeas/sdks-service/server"
)

func main() {
	app := server.New("testing")

	app.WithConfigMap("config-1").WithInMemoryConfiguration("default", map[string]string{
		"test": "value",
	}).WithEnvironmentVariableConfiguration("DOTNET_")

	app.WithLivenessHealthCheck(http.NewHttpHealthCheck("https://google.com"))

	app.ConfigureHandlers(func(f *fiber.App, configuration func() *configuration.ConfigurationRoot) {
		f.Get("/", func(c *fiber.Ctx) error {
			value := configuration().GetStringValueOrDefault("log.level", "not set")
			return c.SendString(fmt.Sprintf("Hello, World 👋! Log Level is: %s", value))
		})
	})

	app.Run()
}
