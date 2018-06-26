package bpq

import (
  "fmt"
)

type ringBufferEntry struct {
  value    interface{}
  priority int64
  inUse    bool
}

type bpqRingBuffer struct {
  entries              []ringBufferEntry
  startIndex, endIndex int
  compareFunc          func(int64, int64) bool
  queueType            QueueType
}

func (e ringBufferEntry) String() string {
  if e.inUse {
    return fmt.Sprintf("{ Value %v, Priority %v }", e.value, e.priority)
  } else {
    return fmt.Sprintf("{}")
  }
}

func (bpq bpqRingBuffer) String() string {
  return fmt.Sprintf("BPQ Ring Buffer: Start %v, End %v, Entries: %v", bpq.startIndex, bpq.endIndex, bpq.entries)
}

func makeRingBuffer(capacity int, queueType QueueType) *bpqRingBuffer {
  var fn func(int64, int64) bool
  if queueType == MaxQueue {
    fn = func(a int64, b int64) bool {
      return a > b
    }
  } else {
    fn = func(a int64, b int64) bool {
      return a < b
    }
  }

  result := bpqRingBuffer{make([]ringBufferEntry, capacity),
    0, 0, fn, queueType}

  for i := 0; i < capacity; i++ {
    result.entries[i] = ringBufferEntry{nil, 0, false}
  }

  return &result
}

func (bpq *bpqRingBuffer) Capacity() int {
  return len(bpq.entries)
}

func (bpq *bpqRingBuffer) QueueType() QueueType {
  return bpq.queueType
}

func (bpq *bpqRingBuffer) Push(item interface{}, priority int64) bool {
  //defer fmt.Printf("Post-Push: %v", bpq)

  if bpq.entries[bpq.endIndex].inUse && bpq.compareFunc(bpq.entries[bpq.endIndex].priority, priority) {
    // We can't insert
    return false
  }

  index := bpq.endIndex
  bpq.entries[index].value = item
  bpq.entries[index].priority = priority
  bpq.entries[index].inUse = true

  nextIndex := (bpq.endIndex + 1) % len(bpq.entries)
  if nextIndex != bpq.startIndex {
    bpq.endIndex = nextIndex
  }

  // Pull it backwards until it's at the right place
  for index != bpq.startIndex && bpq.entries[index].inUse {
    var prevIndex int
    if index == 0 {
      prevIndex = len(bpq.entries) - 1
    } else {
      prevIndex = index - 1
    }

    if bpq.compareFunc(bpq.entries[index].priority, bpq.entries[prevIndex].priority) {
      bpq.entries[prevIndex], bpq.entries[index] = bpq.entries[index], bpq.entries[prevIndex]

      index = prevIndex
    } else {
      break
    }
  }

  return true
}

func (bpq *bpqRingBuffer) Pop() (interface{}, error) {
  if bpq.entries[bpq.startIndex].inUse == false {
    return nil, NoElementsError
  }

  result := bpq.entries[bpq.startIndex].value
  bpq.entries[bpq.startIndex].inUse = false

  bpq.startIndex = bpq.startIndex + 1
  if bpq.startIndex == len(bpq.entries) {
    bpq.startIndex = 0
  }

  return result, nil
}
