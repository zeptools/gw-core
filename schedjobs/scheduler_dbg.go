//go:build debug

package schedjobs

import (
	"log"
	"time"
)

func (s *Scheduler) runOneTimeJobs(now time.Time) {
	key := now.Unix() / 60
	s.mu.Lock()
	jobs := s.oneTimeJobs[key]
	delete(s.oneTimeJobs, key)
	log.Println("[DEBUG] runOneTimeJobs(now) with key: ", key)
	s.mu.Unlock()
	for _, job := range jobs {
		s.runOneTimeJob(job)
	}
}

func (s *Scheduler) runOneTimeJob(job *OneTimeJob) {
	log.Println("[DEBUG] runOneTimeJob() called")
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := job.Task()
		if job.OnFinished != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("[PANIC] Recovered in job.OnFinished:", r)
					}
				}()
				job.OnFinished(err)
			}()
		}
		if s.OnOneTimeJobFinished != nil {
			s.OnOneTimeJobFinished(job, err)
		}
	}()
}

func (s *Scheduler) runCronJobs(now time.Time) {
	log.Println("[DEBUG] runCronJobs called at", now)
	s.mu.Lock()
	log.Println("[DEBUG] total cron jobs:", len(s.cronJobs))
	// Copy values to a slice so we can unlock early
	jobs := make([]*CronJob, 0, len(s.cronJobs))
	for _, job := range s.cronJobs {
		jobs = append(jobs, job)
	}
	log.Printf("[DEBUG] %d cronjobs copied", len(jobs))
	s.mu.Unlock()
	for _, job := range jobs {
		log.Println("[DEBUG] matching cron job spec for ", job.ID)
		if job.Matches(now) {
			log.Println("[DEBUG] cron job spec MATCHED for ", job.ID)
			s.runCronJob(job)
		}
	}
}

func (s *Scheduler) runCronJob(job *CronJob) {
	log.Println("[DEBUG] runCronJob() called")
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
