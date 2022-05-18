package healthchecks

import (
	"fmt"
	"time"
)

type JsonTime struct {
	duration time.Duration
}

func NewJsonTime(duration time.Duration) JsonTime {
	return JsonTime{
		duration: duration,
	}
}

func (t JsonTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%dms\"", t.duration.Milliseconds())
	return []byte(stamp), nil
}
