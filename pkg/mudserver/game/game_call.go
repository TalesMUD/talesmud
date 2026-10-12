package game

import (
	"errors"
	"time"
)

// gameJob is one live-ops function to run on the command loop.
type gameJob struct {
	fn   func()
	done chan struct{}
}

func (g *Game) runJob(job gameJob) {
	defer close(job.done)
	if job.fn != nil {
		job.fn()
	}
}

// Call runs fn on the command loop once Run has started.
// Before Run, and in tests that never start the loop, fn runs on the caller.
// Do not call Call from inside a job: the loop would wait on itself.
func (g *Game) Call(fn func()) error {
	if fn == nil {
		return nil
	}
	if g == nil || !g.loopOn.Load() {
		fn()
		return nil
	}
	done := make(chan struct{})
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case g.jobs <- gameJob{fn: fn, done: done}:
	case <-timer.C:
		return errors.New("game loop did not accept the operation")
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(30 * time.Second)
	select {
	case <-done:
		return nil
	case <-timer.C:
		return errors.New("game loop timed out")
	}
}
