/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package worker

import (
	"fmt"
	"sync"

	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

// Pool contains a pool of workers that can each execute a Task
type Pool struct {
	name            string
	workers         []*Worker
	availableWorker chan *Worker
	taskWg          sync.WaitGroup
	wg              sync.WaitGroup
}

// NewPool creates a worker Pool with the given Factory
func NewPool(name string, concurrency uint16) *Pool {
	pool := &Pool{
		name:            name,
		availableWorker: make(chan *Worker, concurrency),
		workers:         make([]*Worker, concurrency),
	}

	// Create the workers
	for i := 0; i < int(concurrency); i++ {
		pool.workers[i] = newWorker(fmt.Sprintf("%s-%d", name, i), pool)
	}

	return pool
}

// Name returns the name of the pool
func (p *Pool) Name() string {
	return p.name
}

// Start starts the pool
func (p *Pool) Start() {
	p.wg.Add(len(p.workers))

	// Start the workers
	for _, w := range p.workers {
		w.Start()
	}
}

// Stop stops the pool and optionally waits until all tasks have completed
func (p *Pool) Stop(wait bool) {
	cliconfig.Config().Logger().Debugf("[%s] Stopping worker pool ...\n", p.name)

	if wait {
		// Wait for all the tasks to complete
		cliconfig.Config().Logger().Debugf("[%s] ... waiting for tasks to complete ...\n", p.name)
		p.taskWg.Wait()
	} else {
		cliconfig.Config().Logger().Debugf("[%s] ... forcing all tasks to stop ...\n", p.name)
	}

	// Shut down the workers
	cliconfig.Config().Logger().Debugf("[%s] ... stopping workers ...\n", p.name)
	for i := 0; i < len(p.workers); i++ {
		w := <-p.availableWorker
		w.Stop()
	}

	// Wait for all of the workers to stop
	p.wg.Wait()
}

// Submit submits a Task for execution
func (p *Pool) Submit(task Task) {
	cliconfig.Config().Logger().Debugf("worker pool.Submit[%s] - waiting for available worker\n", p.name)

	p.taskWg.Add(1)

	// Wait for an available worker
	w := <-p.availableWorker

	cliconfig.Config().Logger().Debugf("worker pool.Submit[%s] - got worker [%s]. Submitting task...\n", p.name, w.Name())

	// Submit the task to the worker
	w.Submit(task)

	cliconfig.Config().Logger().Debugf("worker pool.Submit[%s] - submitted task to worker[%s]\n", p.name, w.Name())
}

// StateChange is invoked when the state of the Worker changes
func (p *Pool) StateChange(w *Worker, state State) {
	switch state {
	case READY:
		p.availableWorker <- w
		break

	case STOPPED:
		cliconfig.Config().Logger().Debugf("...Worker[%s] stopped\n", w.Name())
		p.wg.Done()
		break

	default:
		cliconfig.Config().Logger().Warnf("Unsupported worker state: %d\n", state)
		break
	}
}

// TaskStarted is invoked when the given Worker begins executing the given Task
func (p *Pool) TaskStarted(w *Worker, task Task) {
	// Nothing to do
}

// TaskCompleted is invoked when the given Worker completed executing the given Task
func (p *Pool) TaskCompleted(w *Worker, task Task) {
	p.taskWg.Done()
}
