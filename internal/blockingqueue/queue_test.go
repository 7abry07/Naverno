package blockingqueue_test

import (
	"Naverno/internal/blockingqueue"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	queue := blockingqueue.New([]int{})
	poppedValues := make(chan []int)
	go func(res chan []int) {
		items := []int{}
		for {
			val, ok := queue.Pop()
			if !ok {
				break
			}
			items = append(items, val)
		}
		res <- items
	}(poppedValues)

	queue.Push(5)
	queue.Push(6)
	queue.Stop()

	testTime := time.NewTimer(time.Second * 2)
	select {
	case res := <-poppedValues:
		if len(res) != 2 {
			t.Fatalf("unexpected value count, expected -> %v, got -> %v", 2, len(res))
		}
		if res[0] != 5 || res[1] != 6 {
			t.Errorf("unexpected values, expected -> %v, %v, got -> %v, %v", 5, 6, res[0], res[1])
		}
	case <-testTime.C:
		t.Fatal("test time exceeded")
	}
}
