package healthchecks

import "time"

type HealthCheckRunner struct {
	readinessChecks []HealthCheck
	livenessChecks  []HealthCheck
}

func (runner *HealthCheckRunner) AddLivenessCheck(check HealthCheck) *HealthCheckRunner {
	runner.livenessChecks = append(runner.livenessChecks, check)
	return runner
}

func (runner *HealthCheckRunner) AddReadinessCheck(check HealthCheck) *HealthCheckRunner {
	runner.livenessChecks = append(runner.readinessChecks, check)
	return runner
}

func (runner HealthCheckRunner) RunLivenessChecks() HealthCheckAggregatedResult {
	return runner.run(runner.livenessChecks)
}

func (runner HealthCheckRunner) RunReadinessChecks() HealthCheckAggregatedResult {
	return runner.run(runner.readinessChecks)
}

func (runner HealthCheckRunner) run(checks []HealthCheck) HealthCheckAggregatedResult {
	result := HealthCheckAggregatedResult{
		State:  HealthCheckState_Healthy,
		Checks: []HealthCheckResult{},
	}

	startTime := time.Now()
	for _, check := range checks {
		hcr := check.Check()
		result.Add(hcr)
	}
	result.Duration = NewJsonTime(time.Since(startTime))

	return result
}
