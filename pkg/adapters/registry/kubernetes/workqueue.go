// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package kubernetes

import (
	"sync"

	"k8s.io/client-go/pkg/api/meta"
)

const (
	workqueueSize = 1024
)

// functions corresponding to the cache event handlers.
type addFunc func(obj interface{})
type updateFunc func(oldObj, newObj interface{})
type deleteFunc func(obj interface{})

// workqueue enqueues and processes event callbacks from cache controllers.
type workqueue struct {
	workChan chan func()
	stopChan chan struct{}

	active bool
	mutex  sync.Mutex
}

// newWorkqueue creates a new workqueue.
func newWorkqueue() *workqueue {
	return &workqueue{
		workChan: make(chan func(), workqueueSize),
		stopChan: make(chan struct{}),
	}
}

// Start launches a worker goroutine to process events from the queue.
func (wq *workqueue) Start() {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()

	if wq.active {
		return
	}
	wq.active = true

	go wq.work()
}

// Stop the worker goroutine from processing events from the queue.
func (wq *workqueue) Stop() {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()

	if !wq.active {
		return
	}
	wq.active = false

	wq.stopChan <- struct{}{}
}

// EnqueueingAddFunc returns an addFunc that enqueues the given addFunc invocation.
func (wq *workqueue) EnqueueingAddFunc(f addFunc) addFunc {
	return func(obj interface{}) {
		wq.enqueueFunc(func() {
			f(obj)
		})
	}
}

// EnqueueingUpdateFunc returns an updateFunc that enqueues the given updateFunc invocation.
// The wrapping updateFunc drops events in which the resource version of the old object is
// the same as the resource version of the new object (e.g., in case of a full cache resync).
func (wq *workqueue) EnqueueingUpdateFunc(f updateFunc) updateFunc {
	return func(oldObj, newObj interface{}) {
		oldMeta, err := meta.Accessor(oldObj)
		if err != nil {
			return
		}

		newMeta, err := meta.Accessor(newObj)
		if err != nil {
			return
		}

		// Drop update in case that resource version hasn't changed (e.g., on full resync)
		if oldMeta.GetResourceVersion() == newMeta.GetResourceVersion() {
			return
		}

		wq.enqueueFunc(func() {
			f(oldObj, newObj)
		})
	}
}

// EnqueueingDeleteFunc returns an deleteFunc that enqueues the given deleteFunc invocation.
func (wq *workqueue) EnqueueingDeleteFunc(f deleteFunc) deleteFunc {
	return func(obj interface{}) {
		wq.enqueueFunc(func() {
			f(obj)
		})
	}
}

// enqueueFunc enqueues the given function for later execution by the worker goroutine
func (wq *workqueue) enqueueFunc(f func()) {
	wq.workChan <- f
}

// work loops and invokes any functions queued for execution.
// It is run by the worker goroutine.
func (wq *workqueue) work() {
	for {
		select {
		case f := <-wq.workChan:
			f()
		case <-wq.stopChan:
			return
		}
	}
}
