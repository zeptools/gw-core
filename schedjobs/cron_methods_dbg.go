//go:build debug

package schedjobs

import (
	"log"
	"time"
)

func (job *CronJob) Matches(now time.Time) bool {
	log.Printf("[DEBUG] Checking match for %s at %v", job.ID, now)
	log.Printf("[DEBUG] Cron spec: Minutes=%v Hours=%v DaysOfMonth=%v Weekdays=%v",
		job.Minutes, job.Hours, job.DaysOfMonth, job.Weekdays,
	)
	if (job.Minutes & (1 << now.Minute())) == 0 {
		log.Println("[DEBUG] Minute mismatch")
		return false
	}
	if (job.Hours & (1 << now.Hour())) == 0 {
		log.Println("[DEBUG] Hour mismatch")
		return false
	}
	if (job.DaysOfMonth & (1 << (now.Day() - 1))) == 0 { // now.Day() = 1..31 -> bit 0 = day 1
		log.Println("[DEBUG] Days of month mismatch")
		return false
	}
	if (job.Weekdays & (1 << now.Weekday())) == 0 {
		log.Println("[DEBUG] weekday mismatch")
		return false
	}
	log.Println("[DEBUG] All fields match")
	return true
}
