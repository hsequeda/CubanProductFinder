package storeClient

import (
	"context"
	"sync"
)

type Worker interface {
	GetArgs() context.Context
	Task(ctx context.Context)
}

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
