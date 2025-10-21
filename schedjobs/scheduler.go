package schedjobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	oneTimeJobs map[int64][]*OneTimeJob
	cronJobs    []*CronJob
	mu          sync.Mutex
	wg          sync.WaitGroup
	cancel      context.CancelFunc
	// Default Callbacks
	OnOneTimeJobAdded    func(job *OneTimeJob)
	OnCronJobAdded       func(job *CronJob)
	OnOneTimeJobFinished func(job *OneTimeJob, err error)
	OnCronJobFinished    func(job *CronJob, err error)
	OnOneTimeJobDeleted  func(job *OneTimeJob)
	OnCronJobDeleted     func(job *CronJob)
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		oneTimeJobs: make(map[int64][]*OneTimeJob),
		cronJobs:    []*CronJob{},
	}
}

func (s *Scheduler) Start() {
	if s.cancel != nil {
		return // already started
	}
	// new derived context `ctx` from the parent `context.Background()`
	ctx, cancel := context.WithCancel(context.Background()) // With cancel(), it notifies all goroutines waiting on ctx.Done()
	s.cancel = cancel
	go s.loop(ctx)
	log.Println("[INFO] job scheduler started")
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait() // wait for running tasks
	log.Println("[INFO] job scheduler stopped")
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		now := time.Now()
		s.runOneTimeJobs(now)
		s.runCronJobs(now)

		select {
		case <-ticker.C:
			// continue for-loop
		case <-ctx.Done():
			return
		}
	}
}

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
	s.mu.Lock()
	jobs := append([]*CronJob(nil), s.cronJobs...) // copy jobs so unlocking early is possible
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

func (s *Scheduler) AddOneTimeJob(job *OneTimeJob) error {
	now := time.Now()
	margin := 30 * time.Second
	if job.ExecTime.Before(now.Add(margin)) {
		return fmt.Errorf(
			"cannot schedule job %s too close or in the past (ExecTime: %s, now: %s)",
			job.ID, job.ExecTime, now,
		)
	}
	// Round up to the next minute if ExecTime has seconds/nanoseconds
	regTime := job.ExecTime
	if regTime.Second() > 0 || regTime.Nanosecond() > 0 {
		regTime = regTime.Truncate(time.Minute).Add(time.Minute)
	}
	key := regTime.Unix() / 60
	s.mu.Lock()
	if s.oneTimeJobs == nil {
		s.oneTimeJobs = make(map[int64][]*OneTimeJob) // safety net
	}
	s.oneTimeJobs[key] = append(s.oneTimeJobs[key], job) // to make this safer?
	s.mu.Unlock()
	if job.OnAdded != nil { // Job-specific callback
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("[PANIC] Recovered in job.OnAdded:", r)
				}
			}()
			job.OnAdded()
		}()
	}
	if s.OnOneTimeJobAdded != nil { // Scheduler-level default callback
		s.OnOneTimeJobAdded(job)
	}
	return nil
}

func (s *Scheduler) AddCronJob(job *CronJob) {
	s.mu.Lock()
	s.cronJobs = append(s.cronJobs, job)
	s.mu.Unlock()
	if job.OnAdded != nil { // Job-specific callback
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("[PANIC] Recovered in job.OnAdded:", r)
				}
			}()
			job.OnAdded()
		}()
	}
	if s.OnCronJobAdded != nil { // Scheduler-level default callback
		s.OnCronJobAdded(job)
	}
}

// GetOneTimeJobs returns a copy of all pending one-time jobs, keyed by their scheduled minute-level timestamp.
func (s *Scheduler) GetOneTimeJobs() map[int64][]*OneTimeJob {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[int64][]*OneTimeJob, len(s.oneTimeJobs))
	for key, jobs := range s.oneTimeJobs {
		result[key] = append([]*OneTimeJob(nil), jobs...) // copy slice to avoid external mutation
	}
	return result
}

// GetCronJobs returns a copy of all registered cron jobs
func (s *Scheduler) GetCronJobs() []*CronJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Make a shallow copy of the slice to avoid external mutation
	jobs := append([]*CronJob(nil), s.cronJobs...)
	return jobs
}

// DeleteOneTimeJob - Delete a job
func (s *Scheduler) DeleteOneTimeJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, jobs := range s.oneTimeJobs {
		filtered := jobs[:0]
		for _, job := range jobs {
			if job.ID == jobID {
				if s.OnOneTimeJobDeleted != nil {
					s.OnOneTimeJobDeleted(job)
				}
			} else {
				filtered = append(filtered, job)
			}
		}
		if len(filtered) == 0 {
			delete(s.oneTimeJobs, key)
		} else {
			s.oneTimeJobs[key] = filtered
		}
	}
}

// DeleteCronJob removes a cron job by its ID
func (s *Scheduler) DeleteCronJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newJobs := s.cronJobs[:0] // reuse underlying array
	for _, job := range s.cronJobs {
		if job.ID != jobID {
			newJobs = append(newJobs, job)
		} else if s.OnCronJobDeleted != nil {
			// trigger global delete callback
			s.OnCronJobDeleted(job)
		}
	}
	s.cronJobs = newJobs
}
