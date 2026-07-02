package clock

import "time"

// Clock is an interface that provides the current time.
type Clock interface{ Now() time.Time }

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

func System() Clock { return systemClock{} }
