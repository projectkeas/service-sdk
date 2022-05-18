package healthchecks

type HealthCheck interface {
	Check() HealthCheckResult
}
