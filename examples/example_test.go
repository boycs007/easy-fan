package examples

import (
	"context"
	"fmt"
	"testing"
	"time"

	fan "github.com/boycs007/easy-fan"
)

// Example demonstrates basic usage of the Fan framework
func Example_simple() {
	ctx := context.Background()
	tasks := []int{1, 2, 3, 4, 5}

	processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
		out := make(chan fan.UnitRet[string])
		go func() {
			defer close(out)
			for num := range inChan {
				out <- fan.UnitRet[string]{
					Item: fmt.Sprintf("Processed: %d", num),
					Err:  nil,
				}
			}
		}()
		return out
	})

	results := processor.WithHandlerNum(3).Do()
	for _, result := range results {
		fmt.Println(result.Item)
	}
	// Output:
	// Processed: 1
	// Processed: 2
	// Processed: 3
	// Processed: 4
	// Processed: 5
}

// Example_error demonstrates error handling in the Fan framework
func Example_error() {
	ctx := context.Background()
	tasks := []int{1, 2, 3, 4, 5}

	processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
		out := make(chan fan.UnitRet[string])
		go func() {
			defer close(out)
			for num := range inChan {
				if num%2 == 0 {
					out <- fan.UnitRet[string]{
						Item: fmt.Sprintf("Processed: %d", num),
						Err:  nil,
					}
				} else {
					out <- fan.UnitRet[string]{
						Item: "",
						Err:  fmt.Errorf("error processing odd number: %d", num),
					}
				}
			}
		}()
		return out
	})

	results := processor.WithHandlerNum(2).Do()
	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("Error: %v\n", result.Err)
		} else {
			fmt.Println(result.Item)
		}
	}
	// Output:
	// Error: error processing odd number: 1
	// Processed: 2
	// Error: error processing odd number: 3
	// Processed: 4
	// Error: error processing odd number: 5
}

// Example_timeout demonstrates context timeout handling
func Example_timeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	tasks := []int{1, 2, 3, 4, 5}

	processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
		out := make(chan fan.UnitRet[string])
		go func() {
			defer close(out)
			for num := range inChan {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(50 * time.Millisecond) // Simulate work
					out <- fan.UnitRet[string]{
						Item: fmt.Sprintf("Processed: %d", num),
						Err:  nil,
					}
				}
			}
		}()
		return out
	})

	results := processor.WithHandlerNum(2).Do()
	fmt.Printf("Processed %d items before timeout\n", len(results))
	// Output:
	// Processed 2 items before timeout
}

// Example_cache demonstrates the use of merge cache
func Example_cache() {
	ctx := context.Background()
	tasks := []int{1, 2, 3, 4, 5}

	processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
		out := make(chan fan.UnitRet[string])
		go func() {
			defer close(out)
			for num := range inChan {
				time.Sleep(10 * time.Millisecond) // Simulate work
				out <- fan.UnitRet[string]{
					Item: fmt.Sprintf("Processed: %d", num),
					Err:  nil,
				}
			}
		}()
		return out
	})

	results := processor.WithHandlerNum(3).WithMergeCache(10).Do()
	fmt.Printf("Successfully processed %d items with cache\n", len(results))
	// Output:
	// Successfully processed 5 items with cache
}

// TestConcurrentProcessing tests concurrent processing with different handler counts
func TestConcurrentProcessing(t *testing.T) {
	ctx := context.Background()
	tasks := make([]int, 100)
	for i := range tasks {
		tasks[i] = i + 1
	}

	testCases := []struct {
		name          string
		handlerCount  int
		expectedCount int
	}{
		{"SingleHandler", 1, 100},
		{"MultipleHandlers", 5, 100},
		{"ExcessiveHandlers", 200, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := fan.GetProcessor(ctx, tasks, func(ctx context.Context, inChan <-chan int) <-chan fan.UnitRet[string] {
				out := make(chan fan.UnitRet[string])
				go func() {
					defer close(out)
					for num := range inChan {
						out <- fan.UnitRet[string]{
							Item: fmt.Sprintf("Processed: %d", num),
							Err:  nil,
						}
					}
				}()
				return out
			})

			results := processor.WithHandlerNum(tc.handlerCount).Do()
			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}
		})
	}
}
