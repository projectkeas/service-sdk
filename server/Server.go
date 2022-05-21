package server

import (
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks"
	log "github.com/projectkeas/sdks-service/logger"
)

type FiberAppFunc func(app *fiber.App, configurationAccessor func() *configuration.ConfigurationRoot)

type Server struct {
	AppName string

	configMaps             []string
	configuration          configuration.ConfigurationRoot
	secrets                []string
	configurationProviders []configuration.ConfigurationProvider
	handlerConfig          FiberAppFunc
	healthCheckRunner      *healthchecks.HealthCheckRunner
}

func New(appName string) Server {
	server := Server{AppName: appName, healthCheckRunner: &healthchecks.HealthCheckRunner{}}
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

func (server *Server) WithInMemoryConfiguration(name string, data map[string]string) *Server {
	return server.WithConfigurationProvider(*configuration.NewInMemoryConfigurationProvider(name, data))
}

func (server *Server) WithEnvironmentVariableConfiguration(prefix string) *Server {
	return server.WithConfigurationProvider(*configuration.NewEnvironmentConfigurationProvider(prefix))
}

func (server *Server) WithConfigurationProvider(provider configuration.ConfigurationProvider) *Server {
	server.configurationProviders = append(server.configurationProviders, provider)
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

func (server *Server) WithReadinessHealthCheck(healthCheck healthchecks.HealthCheck) *Server {
	server.healthCheckRunner.AddLivenessCheck(healthCheck)
	return server
}

func (server *Server) WithLivenessHealthCheck(healthCheck healthchecks.HealthCheck) *Server {
	server.healthCheckRunner.AddLivenessCheck(healthCheck)
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

	for _, provider := range server.configurationProviders {
		builder.AddConfigurationProvider(provider)
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
			result = server.healthCheckRunner.RunReadinessChecks()
		default:
			result = server.healthCheckRunner.RunLivenessChecks()
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
