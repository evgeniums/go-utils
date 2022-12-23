package background_worker

import (
	system_context "context"
	"sync"
	"time"

	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-condchan"
)

type JobRunner func()
type JobStopper func()

type BackgroundWorker struct {
	logger.WithLoggerBase

	Period int

	CondChan *condchan.CondChan
	Finished chan bool
	Stopped  bool
	Running  bool

	RunJob  JobRunner
	StopJob JobStopper
}

type WithBackgroundWorker interface {
	Worker() *BackgroundWorker
}

type WithBackgroundWorkerBase struct {
	WorkerInterface *BackgroundWorker
}

func (w *WithBackgroundWorkerBase) Worker() *BackgroundWorker {
	return w.WorkerInterface
}

func New(log logger.Logger, jobRunner JobRunner, period int) *BackgroundWorker {
	b := &BackgroundWorker{RunJob: jobRunner, Period: period}
	b.SetLogger(log)
	b.CondChan = condchan.New(&sync.Mutex{})
	b.Finished = make(chan bool, 1)
	return b
}

func (w *BackgroundWorker) Run() {

	w.Running = true
	w.Stopped = false

	// run in go routine
	go func() {
		w.RunJob()
		if w.Stopped {
			w.Logger().Debug("Background worker: stopped after first run")
			w.Finished <- true
		}
		for {

			timeoutChan := time.After(time.Second * time.Duration(w.Period+1))
			br := new(bool)
			*br = false

			w.CondChan.L.Lock()
			w.CondChan.Select(func(c <-chan struct{}) {
				select {
				case <-c:
					w.Logger().Debug("Background worker: signal received")
					*br = true
				case <-timeoutChan:
					if !w.Stopped {
						// w.Log.LogDebug("Background worker: run job")
						w.RunJob()
					}
				}
			})
			w.CondChan.L.Unlock()

			if w.Stopped || *br {
				w.Logger().Debug("Background worker: break cycle")
				break
			}
		}
		w.Finished <- true
	}()
}

func (w *BackgroundWorker) Stop() {
	w.Logger().Info("Background worker: stopping...")
	if !w.Running {
		w.Logger().Info("Background worker: not running, quit")
		return
	}
	w.Stopped = true
	if w.StopJob != nil {
		w.StopJob()
	}
	w.CondChan.Broadcast()
	<-w.Finished
	w.Running = false
	w.Logger().Info("Background worker: finished")
}

func (w *BackgroundWorker) Shutdown(ctx system_context.Context) error {
	w.Stop()
	return nil
}
