package reaper

import "time"

type Reaper struct {
	Lifetime time.Duration
	Target   Target
}
