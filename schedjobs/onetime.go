package schedjobs

import "time"

type OneTimeJob struct {
	ID       string
	ExecTime time.Time
	Task     func() error
	OnFinish func(error)
}
