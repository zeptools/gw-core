package schedjobs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Scheduler struct {
	oneTimeJobs        map[int64][]*OneTimeJob
	cronJobs           []*CronJob
	mu                 sync.Mutex
	wg                 sync.WaitGroup
	cancel             context.CancelFunc
	OnAddOneTimeJob    func(job *OneTimeJob)
	OnAddCronJob       func(job *CronJob)
	OnDeleteOneTimeJob func(job *OneTimeJob)
	OnDeleteCronJob    func(job *CronJob)
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
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait() // wait for running tasks
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
		s.runAsync(job.Task, job.OnFinish)
	}
}

func (s *Scheduler) runCronJobs(now time.Time) {
	s.mu.Lock()
	jobs := append([]*CronJob(nil), s.cronJobs...) // copy jobs so unlocking early is possible
	s.mu.Unlock()
	for _, job := range jobs {
		if job.Matches(now) {
			s.runAsync(job.Task, job.OnFinish)
		}
	}
}

func (s *Scheduler) runAsync(task func() error, finish func(error)) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := task()
		if finish != nil {
			finish(err)
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
	if s.OnAddOneTimeJob != nil {
		s.OnAddOneTimeJob(job)
	}
	return nil
}

func (s *Scheduler) AddCronJob(job *CronJob) {
	s.mu.Lock()
	s.cronJobs = append(s.cronJobs, job)
	s.mu.Unlock()
	if s.OnAddCronJob != nil {
		s.OnAddCronJob(job)
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
				if s.OnDeleteOneTimeJob != nil {
					s.OnDeleteOneTimeJob(job)
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
		} else if s.OnDeleteCronJob != nil {
			// trigger global delete callback
			s.OnDeleteCronJob(job)
		}
	}
	s.cronJobs = newJobs
}
