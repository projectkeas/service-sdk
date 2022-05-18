package healthchecks

type HealthCheckAggregatedResult struct {
	State    HealthCheckState    `json:"state"`
	Duration JsonTime            `json:"duration"`
	Checks   []HealthCheckResult `json:"checks"`
}

func (hcar *HealthCheckAggregatedResult) Add(hcr HealthCheckResult) {
	hcar.Checks = append(hcar.Checks, hcr)
	if hcr.State.IsLessHealthierThan(hcar.State) {
		hcar.State = hcr.State
	}
}
