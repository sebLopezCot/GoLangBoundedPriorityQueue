package bpq

import (
  "fmt"
  "math/rand"
  "sort"
  "testing"
)

func TestCapacity(t *testing.T) {
  queue1 := BPQWithCapacity(10, MinQueue)
  queue2 := BPQWithCapacity(20, MinQueue)
  queue3 := BPQWithCapacity(100, MinQueue)

  if queue1.Capacity() != 10 {
    t.Error("Capacity set to 10, was not 10")
  }

  if queue2.Capacity() != 20 {
    t.Error("Capacity set to 20, was not 20")
  }

  if queue3.Capacity() != 100 {
    t.Error("Capacity set to 100, was not 100")
  }
}

// Tests for empty pops

func TestEmptyPopWithRingBuffer(t *testing.T) {
  testEmptyPopWithBPQ(BPQWithCapacity(maxRingBufferSize, MinQueue), t)
}

func TestEmptyPopWithBoundedHeap(t *testing.T) {
  testEmptyPopWithBPQ(BPQWithCapacity(maxRingBufferSize+1, MinQueue), t)
}

func testEmptyPopWithBPQ(queue *BPQ, t *testing.T) {
  r, err := queue.Pop()
  v := r.Value

  if v != nil {
    t.Error("Expected an nil value when popping an empty queue")
  }

  if err == nil {
    t.Error("Expected an error value when popping an empty queue")
  }
}

// Test for simple push/pop

func TestSimplePushPopWithRingBuffer(t *testing.T) {
  testSimplePushPopWithQueue(BPQWithCapacity(maxRingBufferSize, MinQueue), t)
}

func TestSimplePushPopWithBoundedHeap(t *testing.T) {
  testSimplePushPopWithQueue(BPQWithCapacity(maxRingBufferSize+1, MinQueue), t)
}

func testSimplePushPopWithQueue(queue *BPQ, t *testing.T) {
  queue.Push(1, 10)
  r, err := queue.Pop()

  if err != nil {
    t.Error("Unexpected error value when popping a non-empty queue")
  }

  v := r.Value

  if v == nil {
    t.Error("Recieved an unexpected empty value")
  }

  c, suc := v.(int)

  if !suc || c != 1 {
    t.Error(fmt.Sprintf("Expected integer value 1 in reply but got %v", c))
  }
}

// Test for push/pop/pop

func TestPushDoublePopWithRingBuffer(t *testing.T) {
  testPushDoublePopWithQueue(BPQWithCapacity(maxRingBufferSize, MinQueue), t)
}

func TestPushDoublePopWithBoundedHeap(t *testing.T) {
  testPushDoublePopWithQueue(BPQWithCapacity(maxRingBufferSize+1, MinQueue), t)
}

func testPushDoublePopWithQueue(queue *BPQ, t *testing.T) {
  queue.Push(1, 10)
  queue.Pop()
  r, err := queue.Pop()
  if err == nil {
    t.Error("Expected error value when popping an empty queue")
  }

  v := r.Value

  if v != nil {
    t.Error("Did not expect a value when popping an empty queue")
  }
}

// Test for priority ordering

func TestMinPriorityOrderingWithRingBuffer(t *testing.T) {
  testPriorityOrderingWithQueue(BPQWithCapacity(maxRingBufferSize, MinQueue), t)
}

func TestMaxPriorityOrderingWithRingBuffer(t *testing.T) {
  testPriorityOrderingWithQueue(BPQWithCapacity(maxRingBufferSize, MaxQueue), t)
}

func TestMinPriorityOrderingWithBoundedHeap(t *testing.T) {
  testPriorityOrderingWithQueue(BPQWithCapacity(maxRingBufferSize+1, MinQueue), t)
}

func TestMaxPriorityOrderingWithBoundedHeap(t *testing.T) {
  testPriorityOrderingWithQueue(BPQWithCapacity(maxRingBufferSize+1, MaxQueue), t)
}

func testPriorityOrderingWithQueue(queue *BPQ, t *testing.T) {
  cap := queue.Capacity()
  for i := 0; i < cap; i++ {
    if queue.queueType == MaxQueue {
      queue.Push(i, float64(i))
    } else {
      queue.Push(i, float64(cap-1-i))
    }
  }

  badOrder := false
  for i := 0; i < cap; i++ {
    expectedVal := cap - 1 - i
    r, _ := queue.Pop()
    v := r.Value
    if v != expectedVal {
      badOrder = true
      fmt.Printf("Item %v = %v, but expected %v\n", i, v, expectedVal)
    }
  }

  if badOrder {
    t.Errorf("Unexpected ordering for %v", queue.queueType)
  }
}

// Test for over-fill
// Overfill tests currently only test the ring buffer

func TestOverFill(t *testing.T) {
  queue := BPQWithCapacity(5, MinQueue)

  queue.Push(1, 10)
  queue.Push(2, 11)
  queue.Push(3, 20)
  queue.Push(4, 30)
  queue.Push(5, 40)
  queue.Push(6, 50)
  queue.Push(7, 5)
  queue.Push(8, 1)
  queue.Push(9, 50)
  queue.Push(10, 35)

  r1, _ := queue.Pop()
  v1 := r1.Value
  r2, _ := queue.Pop()
  v2 := r2.Value
  r3, _ := queue.Pop()
  v3 := r3.Value
  r4, _ := queue.Pop()
  v4 := r4.Value
  r5, _ := queue.Pop()
  v5 := r5.Value
  r6, err := queue.Pop()
  v6 := r6.Value

  // We should not get a v6
  if v6 != nil || err == nil {
    t.Error("Got more than 5 results!")
  }

  if (v1 == nil || v1.(int) != 8) ||
    (v2 == nil || v2.(int) != 7) ||
    (v3 == nil || v3.(int) != 1) ||
    (v4 == nil || v5.(int) != 3) {
    t.Errorf("Incorrect priority ordering, got %v, %v, %v, %v, %v",
      v1, v2, v3, v4, v5)
  }
}

//
// Psuedo-random tests (with fixed seed)
//

type Entry struct {
  value    int
  priority float64
}

type Entries []Entry

func (es Entries) Len() int {
  return len(es)
}

func (es Entries) Swap(i, j int) {
  es[i], es[j] = es[j], es[i]
}

func (es Entries) Less(i, j int) bool {
  return es[i].priority < es[j].priority
}

func (es Entries) ContainsPriority(priority float64) bool {
  for _, v := range es {
    if v.priority == priority {
      return true
    }
  }

  return false
}

func TestRandomInsertAndPopWithRingBuffer(t *testing.T) {
  testRandomInsertAndPopWithQueue(BPQWithCapacity(maxRingBufferSize, MinQueue),
    maxRingBufferSize, t)
}

func TestRandomInsertAndPopWithBoundedHeap(t *testing.T) {
  testRandomInsertAndPopWithQueue(BPQWithCapacity(100*maxRingBufferSize, MinQueue),
    100*maxRingBufferSize, t)
}

func testRandomInsertAndPopWithQueue(queue *BPQ, max int, t *testing.T) {
  // Use a determined seed for reproducability
  rand.Seed(123456)

  // Generate a random list of entries and priorities

  randomItems := make(Entries, max)
  for i := 0; i < max; i++ {
    var item Entry = Entry{rand.Int(), float64(rand.Int63())}

    // We need to ensure each item has a distinct priority; our priority sort is
    // not stable in any sense
    for {
      if randomItems.ContainsPriority(item.priority) {
        item = Entry{rand.Int() % 100, float64(rand.Int63() % 100)}
      } else {
        break
      }
    }

    randomItems[i] = item
    queue.Push(item.value, item.priority)
  }

  sort.Sort(randomItems)

  for _, entry := range randomItems {
    r, _ := queue.Pop()
    v := r.Value

    if entry.value != v.(int) {
      t.Errorf("Not the right value, got %v but wanted %v", v.(int),
        entry.value)
    }
  }
}

//
// Benchmarks
//

func BenchmarkBPQRingBuffer(b *testing.B) {
  benchmarkBPQ(BPQWithCapacity(maxRingBufferSize, MinQueue), b)
}

func BenchmarkBPQBoundedHeap(b *testing.B) {
  benchmarkBPQ(BPQWithCapacity(100*maxRingBufferSize, MinQueue), b)
}

func benchmarkBPQ(queue *BPQ, b *testing.B) {
  var max = queue.Capacity()

  for n := 0; n < b.N; n++ {
    for i := 0; i < max; i++ {
      queue.Push(i, float64(i%10))
    }

    for i := 0; i < max; i++ {
      queue.Pop()
    }
  }
}
