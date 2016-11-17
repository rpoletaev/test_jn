package ttl

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

//types to implementation container/heap
//container sorted by expiration

//ExpiringSign key and time of expiration
type ExpiringSign struct {
	Expiration int64
	Key        string
}

type expSignHeap []ExpiringSign

func (h expSignHeap) Len() int {

	return len(h)
}

func (h expSignHeap) Less(i, j int) bool {
	return h[i].Expiration < h[j].Expiration
}

func (h expSignHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *expSignHeap) Push(x interface{}) {
	*h = append(*h, x.(ExpiringSign))
}

func (h *expSignHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type ExpQueueOptions struct {
	Period      time.Duration
	DoIfExpired func(string)
	DoIfStoped  func()
}

type ExpiringQueue struct {
	options *ExpQueueOptions
	sync.RWMutex
	heap expSignHeap
	stop chan struct{}
}

func (eq *ExpiringQueue) Stop() {
	eq.stop <- struct{}{}
}

func (eq *ExpiringQueue) Len() int {
	eq.RLock()
	defer eq.RUnlock()
	return eq.heap.Len()
}

func (eq *ExpiringQueue) Push(es ExpiringSign) {
	eq.Lock()
	defer eq.Unlock()
	eq.heap.Push(es)
}

func (eq *ExpiringQueue) Pop() interface{} {
	eq.Lock()
	defer eq.Unlock()
	return eq.heap.Pop()
}

func (eq *ExpiringQueue) del(index int) {
	eq.Lock()
	defer eq.Unlock()
	log.Printf("Delete expired key %s", eq.heap[0].Key)
	heap.Remove(&eq.heap, index)
}

const defaultPeriod = time.Duration(30) * time.Second

func defaultOptions() *ExpQueueOptions {
	return &ExpQueueOptions{
		Period: time.Duration(defaultPeriod),
		DoIfExpired: func(key string) {
			// println(key)
		},
		DoIfStoped: func() {
			println("expiring queue stopped")
		},
	}
}

func CreateExpiringQueue(options *ExpQueueOptions) *ExpiringQueue {
	if options == nil {
		options = defaultOptions()
	}

	if options.Period == 0 {
		options.Period = defaultPeriod
	}

	q := &ExpiringQueue{
		options: options,
		heap:    expSignHeap{},
		stop:    make(chan struct{}),
	}

	heap.Init(&q.heap)

	go func() {
		ticker := time.NewTicker(q.options.Period)
		defer func() {
			ticker.Stop()
			close(q.stop)
			q.options.DoIfStoped()
		}()

		for tick := range ticker.C {
			for q.heap.Len() > 0 {
				if q.heap[0].Expiration <= tick.Unix() {
					q.options.DoIfExpired(q.heap[0].Key)
					q.del(0)
				} else {
					break
				}
			}

			select {
			case <-q.stop:
				return
			default:
				break
			}
		}
	}()
	return q
}
