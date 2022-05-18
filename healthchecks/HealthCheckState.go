package healthchecks

import "fmt"

type HealthCheckState struct {
	slug  string
	value int
}

func (source HealthCheckState) IsHealthierThan(target HealthCheckState) bool {
	return source.value > target.value
}

func (source HealthCheckState) IsLessHealthierThan(target HealthCheckState) bool {
	return source.value < target.value
}

func (source HealthCheckState) Is(target HealthCheckState) bool {
	return source.value == target.value
}

func (source HealthCheckState) String() string {
	return source.slug
}

func (source HealthCheckState) MarshalJSON() ([]byte, error) {
	result := fmt.Sprintf("\"%s\"", source.slug)
	return []byte(result), nil
}

var (
	HealthCheckState_Unknown   = HealthCheckState{slug: "Unknown", value: 0}
	HealthCheckState_Unhealthy = HealthCheckState{slug: "Unhealthy", value: 1}
	HealthCheckState_Degraded  = HealthCheckState{slug: "Degraded", value: 2}
	HealthCheckState_Healthy   = HealthCheckState{slug: "Healthy", value: 3}
)
