package schedjobs

import (
	"log"
	"time"
)

type CronJob struct {
	ID          string
	Minutes     uint64 // 60 bits
	Hours       uint32 // 24 bits
	DaysOfMonth uint32 // 31 bits
	Weekdays    uint8  // 7 bits
	Task        func() error
	// Job-specific callbacks
	OnAdded    func()
	OnFinished func(error)
}

// NewEveryMinEmptyCronJob provides a cronjob matching every minute without a task as a template
// Assign a Task, and Modify its time condition
func NewEveryMinEmptyCronJob(jobID string) *CronJob {
	return &CronJob{
		ID:          jobID,
		Minutes:     AllMinutes,
		Hours:       AllHours,
		DaysOfMonth: AllDaysOfMonth,
		Weekdays:    AllWeekdays,
	}
}

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

const (
	AllMinutes     uint64 = 0xFFFFFFFFFFFFFFF // 60 bits set
	AllHours       uint32 = 0xFFFFFF          // 24 bits set
	AllWeekdays    uint8  = 0b01111111        // sun:0b00000001, mon:0b00000010, ..., fri:0b00100000, sat:0b01000000
	AllDaysOfMonth uint32 = 0x7FFFFFFF        // 31 bits set
)

func BitsFromMinutes(list []int) uint64 {
	var bits uint64
	for _, v := range list {
		if v >= 0 && v < 60 {
			bits |= 1 << v
		}
	}
	return bits
}

func BitsFromHours(list []int) uint32 {
	var bits uint32
	for _, v := range list {
		if v >= 0 && v < 24 {
			bits |= 1 << v
		}
	}
	return bits
}

func BitsFromWeekdays(list []int) uint8 {
	var bits uint8
	for _, v := range list {
		if v >= 0 && v < 7 {
			bits |= 1 << v
		}
	}
	return bits
}

func BitsFromDaysOfMonth(list []int) uint32 {
	var bits uint32
	for _, v := range list {
		if v >= 1 && v <= 31 { // day 1 = bit 0
			bits |= 1 << (v - 1)
		}
	}
	return bits
}
