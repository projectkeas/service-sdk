package server

import (
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks"
	log "github.com/projectkeas/sdks-service/logger"
)

type FiberAppFunc func(app *fiber.App, configurationAccessor func() *configuration.ConfigurationRoot)

type Server struct {
	AppName string

	configMaps    []string
	configuration configuration.ConfigurationRoot
	secrets       []string
	handlerConfig FiberAppFunc
	healthChecks  []healthchecks.HealthCheck
}

func New(appName string) Server {
	server := Server{AppName: appName}
	server.configMaps = []string{}
	server.secrets = []string{}
	return server
}

func (server *Server) GetConfiguration() *configuration.ConfigurationRoot {
	return &server.configuration
}

func (server *Server) ConfigureHandlers(handlerConfig FiberAppFunc) *Server {
	server.handlerConfig = handlerConfig
	return server
}

func (server *Server) WithConfigMap(name string) *Server {
	server.configMaps = append(server.configMaps, name)
	return server
}

func (server *Server) WithSecret(name string) *Server {
	server.secrets = append(server.secrets, name)
	return server
}

func (server *Server) WithHealthCheck(healthCheck healthchecks.HealthCheck) *Server {
	server.healthChecks = append(server.healthChecks, healthCheck)
	return server
}

func (server *Server) Run() {
	runInternal(server, false)
}

func (server *Server) RunDevelopment() {
	runInternal(server, true)
}

func runInternal(server *Server, development bool) {
	server.configuration = setupConfig(server, development, func(config configuration.ConfigurationRoot) {
		log.Initialize(log.Config{
			AppName:       server.AppName,
			LogLevel:      config.GetStringValueOrDefault("log.level", "debug"),
			IsDevelopment: development,
		})
	})

	runServer(server, development)
}

func setupConfig(server *Server, development bool, callback func(configuration.ConfigurationRoot)) configuration.ConfigurationRoot {

	builder := configuration.NewConfigurationBuilder(development)

	for _, name := range server.configMaps {
		builder.AddObservableConfigurationProvider(configuration.NewKubernetesConfigMapConfigurationProvider(name))
	}

	for _, name := range server.secrets {
		builder.AddObservableConfigurationProvider(configuration.NewKubernetesSecretConfigurationProvider(name))
	}

	config := builder.Build(callback)
	callback(config)

	return config
}

func runServer(server *Server, development bool) {
	app := fiber.New(fiber.Config{
		AppName:               server.AppName,
		DisableDefaultDate:    true,
		DisableStartupMessage: !development,
	})

	app.Use(recover.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log.Logger.WithOptions(zap.WithCaller(false)),
		Fields: []string{"status", "method", "url", "ip", "ua", "bytesReceived", "bytesSent", "requestId"},
	}))

	app.Get("/_system/health", func(context *fiber.Ctx) error {

		result := healthchecks.HealthCheckAggregatedResult{
			State: healthchecks.HealthCheckState_Healthy,
		}

		startTime := time.Now()
		for _, check := range server.healthChecks {
			hcr := check.Check()
			result.Add(hcr)
		}
		result.Duration = healthchecks.NewJsonTime(time.Since(startTime))

		context.JSON(result)
		if result.State.Is(healthchecks.HealthCheckState_Healthy) {
			context.SendStatus(200)
		} else {
			context.SendStatus(503)
		}

		return nil
	})

	if server.handlerConfig != nil {
		server.handlerConfig(app, server.GetConfiguration)
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
		log.Logger.Sync()
	}()

	log.Logger.Info("Application starting...")
	err := app.Listen(":5000")
	if err != nil {
		log.Logger.Panic(err.Error())
	}
}
