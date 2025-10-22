//go:build !debug

package schedjobs

import (
	"time"
)

func (job *CronJob) Matches(now time.Time) bool {
	if (job.Minutes & (1 << now.Minute())) == 0 {
		return false
	}
	if (job.Hours & (1 << now.Hour())) == 0 {
		return false
	}
	if (job.DaysOfMonth & (1 << (now.Day() - 1))) == 0 { // now.Day() = 1..31 -> bit 0 = day 1
		return false
	}
	if (job.Weekdays & (1 << now.Weekday())) == 0 {
		return false
	}
	return true
}
