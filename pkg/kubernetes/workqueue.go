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

// AddFunc is the ResourceEventHandler function called when a watched Kubernetes API object is added.
type AddFunc func(obj interface{})

// UpdateFunc is the ResourceEventHandler function called when a watched Kubernetes API object is updated.
type UpdateFunc func(oldObj, newObj interface{})

// DeleteFunc is the ResourceEventHandler function called when a watched Kubernetes API object is deleted.
type DeleteFunc func(obj interface{})

// Workqueue enqueues and processes ResourceEventHandler callbacks from cache controllers.
type Workqueue struct {
	workChan chan func()
	stopChan chan struct{}

	active bool
	mutex  sync.Mutex
}

// NewWorkqueue creates a new workqueue.
func NewWorkqueue() *Workqueue {
	return &Workqueue{
		workChan: make(chan func(), workqueueSize),
		stopChan: make(chan struct{}),
	}
}

// Start launches a worker goroutine to process events from the queue.
func (wq *Workqueue) Start() {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()

	if wq.active {
		return
	}
	wq.active = true

	go wq.work()
}

// Stop the worker goroutine from processing events from the queue.
func (wq *Workqueue) Stop() {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()

	if !wq.active {
		return
	}
	wq.active = false

	wq.stopChan <- struct{}{}
}

// EnqueueingAddFunc returns an AddFunc that enqueues the given AddFunc invocation.
func (wq *Workqueue) EnqueueingAddFunc(f AddFunc) AddFunc {
	return func(obj interface{}) {
		wq.enqueueFunc(func() {
			f(obj)
		})
	}
}

// EnqueueingUpdateFunc returns an UpdateFunc that enqueues the given UpdateFunc invocation.
// The wrapping UpdateFunc drops events in which the resource version of the old object is
// the same as the resource version of the new object (e.g., in case of a full cache resync).
func (wq *Workqueue) EnqueueingUpdateFunc(f UpdateFunc) UpdateFunc {
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

// EnqueueingDeleteFunc returns an DeleteFunc that enqueues the given DeleteFunc invocation.
func (wq *Workqueue) EnqueueingDeleteFunc(f DeleteFunc) DeleteFunc {
	return func(obj interface{}) {
		wq.enqueueFunc(func() {
			f(obj)
		})
	}
}

// enqueueFunc enqueues the given function for later execution by the worker goroutine
func (wq *Workqueue) enqueueFunc(f func()) {
	wq.workChan <- f
}

// work loops and invokes any functions queued for execution.
// It is run by the worker goroutine.
func (wq *Workqueue) work() {
	for {
		select {
		case f := <-wq.workChan:
			f()
		case <-wq.stopChan:
			return
		}
	}
}
