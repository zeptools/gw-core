package schedjobs

import "time"

type OneTimeJob struct {
	ID       string
	ExecTime time.Time
	Task     func() error
	// Job-specific callbacks
	OnAdded    func()
	OnFinished func(error)
}
