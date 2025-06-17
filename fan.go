package fan

import (
	"context"
	"sync"
)

// UnitRet represents a single unit of data returned from a goroutine
// It contains both the processed item and any potential error that occurred during processing
type UnitRet[R any] struct {
	// Item represents the processed data returned by individual goroutines
	Item R
	// Err represents any error that occurred during the processing of the item
	Err error
}

// UnitHandle defines the handler function type for asynchronous processing
// It takes a context and an input channel, and returns an output channel of UnitRet
// The handler should continuously read from the input channel, process each item,
// and send the results to the output channel
type UnitHandle[R, S any] func(ctx context.Context, inChan <-chan S) (outChan <-chan UnitRet[R])

// BatchProcessor defines the interface for parallel task processing
// It provides methods to configure and execute batch processing operations
type BatchProcessor[R, S any] interface {
	// WithMergeCache sets the buffer size for the merge channel
	// This is useful when there's a speed mismatch between producers and consumers
	WithMergeCache(len int) BatchProcessor[R, S]
	// WithHandlerNum sets the number of concurrent handlers
	// This controls how many goroutines will process the input data
	WithHandlerNum(count int) BatchProcessor[R, S]
	// Do execute the batch processing operation and returns the results
	Do() []UnitRet[R]
}

// Fan implements the Fan In/Fan Out pattern for parallel processing
// It distributes work across multiple goroutines and collects their results
type Fan[R, S any] struct {
	ctx          context.Context
	tasksParam   []S
	handle       UnitHandle[R, S]
	cacheLen     int
	handlerCount int
}

// GetProcessor creates a new Fan processor instance
// ctx: context for cancellation and timeout control
// taskDatas: slice of input data to be processed
// h: handler function that processes individual items
func GetProcessor[R, S any](ctx context.Context, taskDatas []S, h UnitHandle[R, S]) BatchProcessor[R, S] {
	return &Fan[R, S]{
		ctx:        ctx,
		tasksParam: taskDatas,
		handle:     h,
		cacheLen:   0,
	}
}

// WithMergeCache sets the buffer size for the merge channel
// len: size of the buffer for the merge channel
func (f *Fan[R, S]) WithMergeCache(len int) BatchProcessor[R, S] {
	f.cacheLen = len
	return f
}

// WithHandlerNum sets the number of concurrent handlers
// count: number of goroutines to process the input data
func (f *Fan[R, S]) WithHandlerNum(count int) BatchProcessor[R, S] {
	f.handlerCount = count
	return f
}

// Do execute the Fan In/Fan Out pattern
// It distributes work across multiple goroutines and collects their results
// Returns a slice of UnitRet containing all processed results
func (f *Fan[R, S]) Do() []UnitRet[R] {
	// fan out: distribute work to multiple handlers
	ch := f.producer()
	chs := make([]<-chan UnitRet[R], 0)

	concurrentNum := f.handlerCount

	paramCount := len(f.tasksParam)
	if concurrentNum <= 0 || paramCount < f.handlerCount {
		concurrentNum = paramCount
	}

	for i := 0; i < concurrentNum; i++ {
		chs = append(chs, f.handle(f.ctx, ch))
	}

	// fan in: collect results from all handlers
	retChan := f.merge(chs...)
	retList := make([]UnitRet[R], 0)
	for ret := range retChan {
		retList = append(retList, ret)
	}
	return retList
}

// producer creates a channel and sends all input data to it
// Returns a channel that will receive all input data
func (f *Fan[R, S]) producer() <-chan S {
	out := make(chan S)
	go func() {
		defer close(out)
		for _, n := range f.tasksParam {
			out <- n
		}
	}()
	return out
}

// mergeChan creates a channel for merging results
// If cacheLen is set, creates a buffered channel
// This is useful when there's a speed mismatch between producers and consumers
func (f *Fan[R, S]) mergeChan() chan UnitRet[R] {
	if f.cacheLen > 0 {
		return make(chan UnitRet[R], f.cacheLen)
	}
	return make(chan UnitRet[R])
}

// merge combines multiple input channels into a single output channel
// It concurrently reads from all input channels and forwards the data to the output channel
// When all input channels are closed, it closes the output channel
func (f *Fan[R, S]) merge(inChannels ...<-chan UnitRet[R]) <-chan UnitRet[R] {
	out := f.mergeChan()
	var wg sync.WaitGroup
	collect := func(in <-chan UnitRet[R]) {
		defer wg.Done()
		for n := range in {
			out <- n
		}
	}
	wg.Add(len(inChannels))

	for _, c := range inChannels {
		vc := c
		go collect(vc)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
