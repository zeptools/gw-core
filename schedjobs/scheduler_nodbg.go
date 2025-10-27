//go:build !debug

package schedjobs

import (
	"time"
)

func (s *Scheduler) runOneTimeJobs(now time.Time) {
	key := now.Unix() / 60
	s.mu.Lock()
	jobs := s.oneTimeJobs[key]
	delete(s.oneTimeJobs, key)
	s.mu.Unlock()
	for _, job := range jobs {
		s.runOneTimeJob(job)
	}
}

func (s *Scheduler) runOneTimeJob(job *OneTimeJob) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := job.Task()
		if job.OnFinished != nil {
			job.OnFinished(err)
		}
		if s.OnOneTimeJobFinished != nil {
			s.OnOneTimeJobFinished(job, err)
		}
	}()
}

func (s *Scheduler) runCronJobs(now time.Time) {
	s.mu.Lock()
	// Copy values to a slice so we can unlock early
	jobs := make([]*CronJob, 0, len(s.cronJobs))
	for _, job := range s.cronJobs {
		jobs = append(jobs, job)
	}
	s.mu.Unlock()
	for _, job := range jobs {
		if job.Matches(now) {
			s.runCronJob(job)
		}
	}
}

func (s *Scheduler) runCronJob(job *CronJob) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := job.Task()
		if job.OnFinished != nil {
			job.OnFinished(err)
		}
		if s.OnCronJobFinished != nil {
			s.OnCronJobFinished(job, err)
		}
	}()
}
