package main

import (
	"fmt"
	"sync"
	"time"
)

// BlockingQueue is a FIFO container with a fixed capacity.
// It  blocks a reader/writer when it is empty/full.
type BlockingQueue struct {
	m        sync.Mutex
	c        sync.Cond
	data     []interface{}
	capacity int
}

// NewBlockingQueue constructs a BlockingQueue with a given capacity.
func NewBlockingQueue(capacity int) *BlockingQueue {
	q := new(BlockingQueue)
	q.c = sync.Cond{L: &q.m}
	q.capacity = capacity
	return q
}

// Put puts an item in the queue and blocks it the queue is full.
func (q *BlockingQueue) Put(item interface{}) {
	q.c.L.Lock()
	defer q.c.L.Unlock()

	for q.isFull() {
		q.c.Wait()
	}
	q.data = append(q.data, item)
	q.c.Signal()
}

// isFull returns true if the capacity is reached.
func (q *BlockingQueue) isFull() bool {
	return len(q.data) == q.capacity
}

// Take takes an item from the queue and blocks if the queue is empty.
func (q *BlockingQueue) Take() interface{} {
	q.c.L.Lock()
	defer q.c.L.Unlock()

	for q.isEmpty() {
		q.c.Wait()
	}
	result := q.data[0]
	q.data = q.data[1:len(q.data)]
	q.c.Signal()
	return result
}

// isEmpty returns true if there are no elements in the queue.
func (q *BlockingQueue) isEmpty() bool {
	return len(q.data) == 0
}

func TestBlockingQueue() {

	bq := NewBlockingQueue(1)
	done := make(chan bool)

	// slow writer
	go func() {
		bq.Put("A")
		time.Sleep(100 * time.Millisecond)
		bq.Put("B")
		time.Sleep(100 * time.Millisecond)
		bq.Put("C")
	}()

	// reader will be blocked
	go func() {
		item := bq.Take()
		fmt.Printf("Got %v\n", item)
		item = bq.Take()
		fmt.Printf("Got %v\n", item)
		item = bq.Take()
		fmt.Printf("Got %v\n", item)
		done <- true
	}()

	// block while done
	<-done
}

// BlockingQueue is a FIFO container with a fixed capacity.
// It  blocks a reader when it is empty and a writer when it is full.
type BlockingQueueChan struct {
	channel chan interface{}
}

// NewBlockingQueue constructs a BlockingQuee with a given capacity.
func NewBlockingQueueChan(capacity int) *BlockingQueueChan {
	q := BlockingQueueChan{make(chan interface{}, capacity)}
	return &q
}

// Put puts an item in the queue and blocks it the queue is full.
func (q *BlockingQueueChan) Put(item interface{}) {
	q.channel <- item
}

// Take takes an item from the queue and blocks if the queue is empty.
func (q *BlockingQueueChan) Take() interface{} {
	return <-q.channel
}

func main() {
	TestBlockingQueue()
	TestChannelBlockingQueue()
}

func TestChannelBlockingQueue() {
	chanBlockingQueue := NewBlockingQueueChan(10)
	done := make(chan bool)
	go func() {
		fmt.Println("Thread 1 is inserting 1")
		chanBlockingQueue.Put(1)
		time.Sleep(2000 * time.Millisecond)
		fmt.Println("Thread 1 is inserting 2")
		chanBlockingQueue.Put(2)
		time.Sleep(2000 * time.Millisecond)
		fmt.Println("Thread 1 is inserting 3")
		chanBlockingQueue.Put(3)
		fmt.Println("Thread 1 is done inserting")
	}()
	go func() {
		fmt.Println("Thread 2 is reading 1st element")
		fmt.Println(chanBlockingQueue.Take())
		fmt.Println("Thread 2 is reading 2nd element")
		fmt.Println(chanBlockingQueue.Take())
		fmt.Println("Thread 2 is reading 3rd element")
		fmt.Println(chanBlockingQueue.Take())
		fmt.Println("Thread 2 is done reading")
		done <- true
	}()
	<-done
}
