package healthchecks

type HealthCheckResult string

const (
	Unknown HealthCheckResult = "Unknown"
	Healthy = "Healthy"
	Degraded = "Degraded"
	Unhealthy = "Unhealthy"
)
