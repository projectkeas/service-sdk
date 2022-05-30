package configHealthCheck

import (
	"time"

	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/healthchecks"
)

type KubernetesConfigMapCheck struct {
	name   string
	config *configuration.ConfigurationRoot
}

type KubernetesSecretCheck struct {
	name   string
	config *configuration.ConfigurationRoot
}

func NewKubernetesConfigMapCheck(name string, config *configuration.ConfigurationRoot) KubernetesConfigMapCheck {
	return KubernetesConfigMapCheck{
		name:   name,
		config: config,
	}
}

func NewKubernetesSecretCheckCheck(name string, config *configuration.ConfigurationRoot) KubernetesConfigMapCheck {
	return KubernetesConfigMapCheck{
		name:   name,
		config: config,
	}
}

func (healthCheck KubernetesConfigMapCheck) Check() healthchecks.HealthCheckResult {
	result := healthchecks.HealthCheckResult{
		Duration: healthchecks.NewJsonTime(0 * time.Millisecond),
		Name:     "ConfigurationCheck",
		State:    healthchecks.HealthCheckState_Unhealthy,
		Data: map[string]string{
			"name": healthCheck.name,
			"type": "ConfigMap",
		},
	}

	for _, provider := range healthCheck.config.Providers {
		cp, ok := provider.(*configuration.KubernetesConfigMapConfigurationProvider)
		if ok && cp.Name() == healthCheck.name {
			if cp.Exists {
				result.State = healthchecks.HealthCheckState_Healthy
			}
			return result
		}
	}

	return result
}

func (healthCheck KubernetesSecretCheck) Check() healthchecks.HealthCheckResult {
	result := healthchecks.HealthCheckResult{
		Duration: healthchecks.NewJsonTime(0 * time.Millisecond),
		Name:     "ConfigurationCheck",
		State:    healthchecks.HealthCheckState_Unhealthy,
		Data: map[string]string{
			"name": healthCheck.name,
			"type": "Secret",
		},
	}

	for _, provider := range healthCheck.config.Providers {
		cp, ok := provider.(*configuration.KubernetesSecretConfigurationProvider)
		if ok && cp.Name() == healthCheck.name {
			if cp.Exists {
				result.State = healthchecks.HealthCheckState_Healthy
			}
			return result
		}
	}

	return result
}
