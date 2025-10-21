package schedjobs

import "time"

type CronJob struct {
	ID          string
	Minutes     uint64 // 60 bits
	Hours       uint32 // 24 bits
	Weekdays    uint8  // 7 bits
	DaysOfMonth uint32 // 31 bits
	Task        func() error
	OnFinish    func(error)
}

func (job *CronJob) Matches(now time.Time) bool {
	if (job.Minutes & (1 << now.Minute())) == 0 {
		return false
	}
	if (job.Hours & (1 << now.Hour())) == 0 {
		return false
	}
	if (job.Weekdays & (1 << now.Weekday())) == 0 {
		return false
	}
	if (job.DaysOfMonth & (1 << (now.Day() - 1))) == 0 { // now.Day() = 1..31 -> bit 0 = day 1
		return false
	}
	return true
}

const (
	AllMinutes     uint64 = 0xFFFFFFFFFFFFFFF // 60 bits set
	AllHours       uint32 = 0xFFFFFF          // 24 bits set
	AllWeekdays    uint8  = 0x7               // 7 bits set
	AllDaysOfMonth uint32 = 0x7FFFFFFF        // 31 bits set
)

func BitsFromList(list []int) uint64 {
	var bits uint64
	for _, v := range list {
		bits |= 1 << v
	}
	return bits
}
