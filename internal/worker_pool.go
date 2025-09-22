package workerpool

import (
	"errors"
	"sync"
)

type Pool interface { 
	Submit(task func()) error
	Stop() error
}

type WorkerPool struct {
	isClosed bool
	mx sync.Mutex
	wg sync.WaitGroup
	chTask chan func()
	chDone chan struct{}
}

func NewWorkerPool(numberOfWorkers, sizeQueue int) *WorkerPool {
	if numberOfWorkers <= 0 || sizeQueue <= 0{
		return nil
	}
	wp := &WorkerPool{
		isClosed: false,
		mx: sync.Mutex{},
		wg: sync.WaitGroup{},
		chTask: make(chan func(), sizeQueue),
		chDone: make(chan struct{}),
	}

	go func(){
		var wg_ sync.WaitGroup
		wg_.Add(numberOfWorkers)
		for i := 0; i < numberOfWorkers; i++ {
			go func(){
				defer wg_.Done()
				for task := range wp.chTask{
					if task != nil{
						task()
					}
				}
			}()
		}
		wg_.Wait()
		close(wp.chDone)
	}()
	return wp
}

func(wp *WorkerPool) Submit(task func()) error {
	if task == nil{
		return errors.New("incorrect task")
	}
	wp.mx.Lock()
	defer wp.mx.Unlock()
	if wp.isClosed{
		return errors.New("worker pool is closed")
	}
	w := func(){
		defer wp.wg.Done()
		task()
	}
    select{
	case wp.chTask <- w:
		wp.wg.Add(1)
	default:
		return errors.New("worker pool overflow")
	}
	return nil
}

func(wp *WorkerPool) Stop() error {
	wp.mx.Lock()
	if wp.isClosed{
		wp.mx.Unlock()
		return errors.New("worker pool is closed")
	}
	wp.isClosed = true
	close(wp.chTask)
	wp.mx.Unlock()
	wp.wg.Wait()
	<-wp.chDone
	return nil
}