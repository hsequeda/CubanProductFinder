package storeClient

import (
	"context"
	"sync"
)

type Worker interface {
	GetArgs() context.Context
	Task(ctx context.Context)
}

// type Worker struct {
// 	Args context.Context
// 	Task func(ctx context.Context)
// }

type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

func NewPool(maxGoroutines int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}

	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				w.Task(w.GetArgs())
			}
			p.wg.Done()
		}()
	}

	return &p
}

func (p *Pool) Run(w Worker) {
	p.work <- w
}

func (p *Pool) Shutdown() {
	close(p.work)
	p.wg.Wait()
}

// func RunWorkers(numWorkers int, do func(interface{}), done chan struct{}, maxChannelSize int) map[int]chan interface{} {
//
// 	workers := make(map[int]chan interface{})
// 	for i := 0; i < numWorkers; i++ {
// 		workerCh := make(chan interface{}, maxChannelSize)
// 		workers[i] = workerCh
//
// 		go func(ch chan interface{}) {
// 			for {
// 				select {
// 				case <-done:
// 					return
// 				case msg := <-ch:
// 					work
					// do(msg)
				// }
			// }
		// }(workerCh)
	// }
	//
	// return workers
// }
