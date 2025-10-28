package schedjobs

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/zeptools/gw-core/svc"
)

type Scheduler struct {
	Ctx         context.Context    // Service Context
	cancel      context.CancelFunc // Service Context CancelFunc
	state       int                // internal service state
	done        chan error         // Shutdown Error Channel
	oneTimeJobs map[int64][]*OneTimeJob
	cronJobs    map[string]*CronJob
	mu          sync.Mutex
	wg          sync.WaitGroup
	// Default Callbacks
	OnOneTimeJobAdded    func(job *OneTimeJob)
	OnCronJobAdded       func(job *CronJob)
	OnOneTimeJobFinished func(job *OneTimeJob, err error)
	OnCronJobFinished    func(job *CronJob, err error)
	OnOneTimeJobDeleted  func(job *OneTimeJob)
	OnCronJobDeleted     func(job *CronJob)
}

func (s *Scheduler) Name() string {
	return "JobScheduler"
}

func NewScheduler(parentCtx context.Context) *Scheduler {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &Scheduler{
		Ctx:         svcCtx,
		cancel:      svcCancel,
		state:       svc.StateREADY,
		done:        make(chan error, 1),
		oneTimeJobs: make(map[int64][]*OneTimeJob),
		cronJobs:    make(map[string]*CronJob),
	}
}

// UseDefaultLoggers set scheduer-level loggers with default ones
func (s *Scheduler) UseDefaultLoggers() {
	s.OnOneTimeJobAdded = func(job *OneTimeJob) {
		log.Printf("[INFO] One-time job added: %s for %v", job.ID, job.ExecTime)
	}
	s.OnCronJobAdded = func(job *CronJob) {
		log.Printf("[INFO] cron job added: %s", job.ID)
	}
	s.OnCronJobFinished = func(job *CronJob, err error) {
		if err == nil {
			log.Printf("[INFO] cron job finished: %s", job.ID)
		} else {
			log.Printf("[INFO] cron job finished: %s with error: %v", job.ID, err)
		}
	}
	s.OnOneTimeJobFinished = func(job *OneTimeJob, err error) {
		if err == nil {
			log.Printf("[INFO] one-time job finished: %s", job.ID)
		} else {
			log.Printf("[INFO] one-time job finished: %s with error: %v", job.ID, err)
		}
	}
}

func (s *Scheduler) Start() error {
	if s.state == svc.StateRUNNING {
		return fmt.Errorf("already started")
	}
	if s.state != svc.StateREADY {
		return fmt.Errorf("cannot start. not ready")
	}
	s.state = svc.StateRUNNING
	log.Println("[INFO][JobScheduler] service started")
	go s.run()
	return nil
}

func (s *Scheduler) Stop() {
	if s.state != svc.StateRUNNING {
		log.Println("[ERROR][JobScheduler] cannot stop. not running")
		return
	}
	s.cancel()
	s.state = svc.StateSTOPPED
	log.Println("[INFO][JobScheduler] service stopped")
}

func (s *Scheduler) Done() <-chan error {
	return s.done
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.Ctx.Done():
			log.Println("[INFO][Scheduler] shutting down...")
			s.wg.Wait()   // wait for all worker goroutines
			s.done <- nil // clean shutdown
			return
		case now := <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC][Scheduler] panic recovered: %v\n%s", r, debug.Stack())
					}
				}()
				s.runOneTimeJobs(now)
				s.runCronJobs(now)
			}()
		}
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

// GetCronJobs returns a copy of all registered cron jobs, keyed by their ID.
func (s *Scheduler) GetCronJobs() map[string]*CronJob {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]*CronJob, len(s.cronJobs))
	for id, job := range s.cronJobs {
		result[id] = job // shallow copy of the pointer; job itself is shared
	}
	return result
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

func (s *Scheduler) AddCronJob(job *CronJob) error {
	s.mu.Lock()
	if s.cronJobs == nil {
		s.cronJobs = make(map[string]*CronJob)
	}
	if _, exists := s.cronJobs[job.ID]; exists {
		return fmt.Errorf("cron job with ID %q already exists", job.ID)
	}
	s.cronJobs[job.ID] = job
	s.mu.Unlock()
	// Job-specific callback
	if job.OnAdded != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("[PANIC] Recovered in job.OnAdded:", r)
				}
			}()
			job.OnAdded()
		}()
	}
	// Scheduler-level default callback
	if s.OnCronJobAdded != nil {
		s.OnCronJobAdded(job)
	}
	return nil
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
	job, exists := s.cronJobs[jobID]
	if !exists {
		s.mu.Unlock()
		return
	}
	delete(s.cronJobs, jobID)
	s.mu.Unlock()
	// trigger global delete callback outside lock
	if s.OnCronJobDeleted != nil {
		s.OnCronJobDeleted(job)
	}
}
