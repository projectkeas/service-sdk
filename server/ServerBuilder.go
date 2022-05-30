package server

import (
	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks"
	"github.com/projectkeas/sdks-service/healthchecks/configHealthCheck"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/opa"
)

type ServerBuilder struct {
	AppName string

	configMaps             []string
	requiredConfigMaps     []string
	secrets                []string
	requiredSecrets        []string
	configurationProviders []configuration.ConfigurationProvider
	handlerConfig          FiberAppFunc
	livenessChecks         []healthchecks.HealthCheck
	readinessChecks        []healthchecks.HealthCheck
	services               map[string]interface{}
}

func New(appName string) *ServerBuilder {
	return &ServerBuilder{
		AppName:  appName,
		services: map[string]interface{}{},
	}
}

func (builder *ServerBuilder) Build() *Server {
	return builder.BuildForDevelopment(false)
}

func (builder *ServerBuilder) BuildForDevelopment(isDevelopment bool) *Server {

	server := newServer(builder.AppName, builder.handlerConfig)
	config := setupConfig(builder, isDevelopment, func(config configuration.ConfigurationRoot) {
		log.Initialize(log.Config{
			AppName:       builder.AppName,
			LogLevel:      config.GetStringValueOrDefault("log.level", "debug"),
			IsDevelopment: isDevelopment,
		})
	})

	// Ensure that we add health checks for the required properties
	for _, key := range builder.requiredConfigMaps {
		builder.WithReadinessHealthCheck(configHealthCheck.NewKubernetesConfigMapCheck(key, config))
	}
	for _, key := range builder.requiredSecrets {
		builder.WithReadinessHealthCheck(configHealthCheck.NewKubernetesSecretCheckCheck(key, config))
	}

	// Register default services
	builder.WithService(configuration.SERVICE_NAME, config)
	builder.WithService(healthchecks.SERVICE_NAME, healthchecks.NewFromHealthChecks(builder.livenessChecks, builder.readinessChecks))
	builder.WithService(opa.SERVICE_NAME, opa.OPAService{})

	for key, svc := range builder.services {
		server.RegisterService(key, svc)
	}

	return &server
}

func (builder *ServerBuilder) ConfigureHandlers(handlerConfig FiberAppFunc) *ServerBuilder {
	builder.handlerConfig = handlerConfig
	return builder
}

func (builder *ServerBuilder) WithInMemoryConfiguration(name string, data map[string]string) *ServerBuilder {
	return builder.WithConfigurationProvider(*configuration.NewInMemoryConfigurationProvider(name, data))
}

func (builder *ServerBuilder) WithEnvironmentVariableConfiguration(prefix string) *ServerBuilder {
	return builder.WithConfigurationProvider(*configuration.NewEnvironmentConfigurationProvider(prefix))
}

func (builder *ServerBuilder) WithConfigurationProvider(provider configuration.ConfigurationProvider) *ServerBuilder {
	builder.configurationProviders = append(builder.configurationProviders, provider)
	return builder
}

func (builder *ServerBuilder) WithConfigMap(name string) *ServerBuilder {
	builder.configMaps = append(builder.configMaps, name)
	return builder
}

func (builder *ServerBuilder) WithRequiredConfigMap(name string) *ServerBuilder {
	builder.requiredConfigMaps = append(builder.requiredConfigMaps, name)
	return builder.WithConfigMap(name)
}

func (builder *ServerBuilder) WithSecret(name string) *ServerBuilder {
	builder.secrets = append(builder.secrets, name)
	return builder
}

func (builder *ServerBuilder) WithRequiredSecret(name string) *ServerBuilder {
	builder.requiredSecrets = append(builder.requiredSecrets, name)
	return builder.WithSecret(name)
}

func (builder *ServerBuilder) WithReadinessHealthCheck(healthCheck healthchecks.HealthCheck) *ServerBuilder {
	builder.readinessChecks = append(builder.readinessChecks, healthCheck)
	return builder
}

func (builder *ServerBuilder) WithLivenessHealthCheck(healthCheck healthchecks.HealthCheck) *ServerBuilder {
	builder.livenessChecks = append(builder.livenessChecks, healthCheck)
	return builder
}

func (builder *ServerBuilder) WithService(name string, service interface{}) *ServerBuilder {
	builder.services[name] = service
	return builder
}

func setupConfig(builder *ServerBuilder, development bool, callback func(configuration.ConfigurationRoot)) *configuration.ConfigurationRoot {

	configurationBuilder := configuration.NewConfigurationBuilder(development)

	for _, name := range builder.configMaps {
		configurationBuilder.AddObservableConfigurationProvider(configuration.NewKubernetesConfigMapConfigurationProvider(name))
	}

	for _, name := range builder.secrets {
		configurationBuilder.AddObservableConfigurationProvider(configuration.NewKubernetesSecretConfigurationProvider(name))
	}

	for _, provider := range builder.configurationProviders {
		configurationBuilder.AddConfigurationProvider(provider)
	}

	config := configurationBuilder.Build(callback)
	callback(*config)

	return config
}
