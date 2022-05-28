package server

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks"
	log "github.com/projectkeas/sdks-service/logger"
)

type FiberAppFunc func(app *fiber.App, server *Server)

type Server struct {
	AppName string

	handlerConfig FiberAppFunc
	services      map[string]interface{}
}

func newServer(appName string, handlerConfig FiberAppFunc) *Server {
	server := &Server{
		AppName:       appName,
		handlerConfig: handlerConfig,
		services:      map[string]interface{}{},
	}
	return server
}

func (server *Server) GetConfiguration() configuration.ConfigurationRoot {
	svc, _ := server.GetService(configuration.SERVICE_NAME)
	return svc.(configuration.ConfigurationRoot)
}

func (server *Server) GetHealthCheckRunner() healthchecks.HealthCheckRunner {
	svc, _ := server.GetService(healthchecks.SERVICE_NAME)
	return svc.(healthchecks.HealthCheckRunner)
}

func (server *Server) RegisterService(name string, service interface{}) {
	_, castSuccessful := service.(Disposable)
	if castSuccessful {
		log.Logger.Info(fmt.Sprintf("Registering service: %s", name))
	}

	server.services[name] = service
}

func (server Server) GetService(name string) (interface{}, error) {
	service, found := server.services[name]

	if found {
		return service, nil
	}

	return nil, fmt.Errorf("unable to locate service '%s'", name)
}

func (server *Server) Run() {
	runServer(server, false)
}

func (server *Server) RunDevelopment() {
	runServer(server, true)
}

func runServer(server *Server, development bool) {
	app := fiber.New(fiber.Config{
		AppName:               server.AppName,
		DisableDefaultDate:    true,
		DisableStartupMessage: !development,
		EnablePrintRoutes:     development,
	})

	app.Use(recover.New())
	app.Use(NewHttpLoggingMiddleware(&LoggingConfig{}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Get("/_system/health/:type?", func(context *fiber.Ctx) error {
		var result healthchecks.HealthCheckAggregatedResult

		switch context.Params("type") {
		case "ready":
			result = server.GetHealthCheckRunner().RunReadinessChecks()
		default:
			result = server.GetHealthCheckRunner().RunLivenessChecks()
		}

		context.JSON(result)
		if result.State.Is(healthchecks.HealthCheckState_Healthy) {
			context.SendStatus(200)
		} else {
			context.SendStatus(503)
		}

		return nil
	})

	if server.handlerConfig != nil {
		server.handlerConfig(app, server)
	}

	// Handle graceful shutdown by proxying with a channel
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt)

	// Use a GoRoutine to monitor the channel and call app.Shutdown
	go func() {
		<-shutdownChannel
		log.Logger.Info("Application stopping...")
		err := app.Shutdown()
		if err != nil {
			log.Logger.Panic(err.Error())
		}
	}()

	log.Logger.Info("Application starting...")
	err := app.Listen(":5000")
	if err != nil {
		log.Logger.Panic(err.Error())
	}

	for key, svc := range server.services {
		disposable, castSuccessful := svc.(Disposable)
		if castSuccessful {
			log.Logger.Info(fmt.Sprintf("Deregistering service: %s", key))
			disposable.Dispose()
		}
	}

	log.Logger.Sync()
}
