package service

import (
	"github.com/coosir/middleman/dispatcher"
	"github.com/coosir/middleman/jobqueue"
	"github.com/coosir/middleman/jobqueue/factory"
	"github.com/coosir/middleman/model"
)

// RunningQueue is an interface of a running queue, which is a job
// queue and its dispatcher combined.
type RunningQueue interface {
	jobqueue.JobQueue
	PollingInterval() uint
	MaxWorkers() uint
	WorkerStats() *dispatcher.Stats
	Deactivate() <-chan struct{}
}

type runningQueue struct {
	jobqueue.JobQueue
	dispatcher dispatcher.Dispatcher
}

func startJobQueue(q *model.Queue) *runningQueue {
	jq := factory.Start(q)
	d := dispatcher.Start(jq, q)
	return &runningQueue{jq, d}
}

func (q *runningQueue) Deactivate() <-chan struct{} {
	deactivated := make(chan struct{})
	go func() {
		<-q.dispatcher.Stop()
		deactivated <- struct{}{}
	}()
	return deactivated
}

func (q *runningQueue) Stop() <-chan struct{} {
	stopped := make(chan struct{})
	go func() {
		<-q.JobQueue.Stop()
		stopped <- struct{}{}
	}()
	return stopped
}

func (q *runningQueue) Push(job jobqueue.IncomingJob) (uint64, error) {
	id, err := q.JobQueue.Push(job)
	q.dispatcher.Ping()
	return id, err
}

func (q *runningQueue) PollingInterval() uint {
	return q.dispatcher.PollingInterval()
}

func (q *runningQueue) MaxWorkers() uint {
	return q.dispatcher.MaxWorkers()
}

func (q *runningQueue) WorkerStats() *dispatcher.Stats {
	if q.IsActive() {
		return q.dispatcher.Stats()
	}
	return &dispatcher.Stats{}
}
