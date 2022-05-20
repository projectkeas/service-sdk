package healthchecks

type HealthCheckResult struct {
	Name     string            `json:"name"`
	State    HealthCheckState  `json:"state"`
	Duration JsonTime          `json:"duration"`
	Data     map[string]string `json:"data"`
}
