package bpq

import (
  "errors"
)

type QueueType int

type QueueItem struct {
  Value    interface{}
  Priority float64
}

const (
  MinQueue QueueType = 1 + iota
  MaxQueue
)

var queueTypes = [...]string{
  "MinQueue",
  "MaxQueue",
}

func (queueType QueueType) String() string {
  return queueTypes[queueType-1]
}

// BPQ represents a bounded min-priority queue
type BPQ interface {
  // Capacity returns the capacity (bounds) on the underlying queue
  Capacity() int
  // Push attempts to enqueue an item to the queue, with the given priority
  // Returns whether it succeded or not (failure is only due to a full queue,
  // where all items are lower priority)
  Push(interface{}, float64) bool
  // Pop attempts to pop an item from the queue. It returns the popped item,
  // or an error. The only error it returns is NoElementsError, indicating
  // that the queue is empty
  Pop() (QueueItem, error)
  // QueueType returns the type of the underlying queue
  QueueType() QueueType
}

// NoElementsError indicates that a Pop was attempted on an empty queue
var NoElementsError error = errors.New("NoElementsError")

const maxRingBufferSize = 128

// BPQWithCapacity creates a new bounded priority queue with the given capacity
func BPQWithCapacity(capacity int, queueType QueueType) BPQ {
  if capacity <= maxRingBufferSize {
    return makeRingBuffer(capacity, queueType)
  } else {
    return makeBoundedHeap(capacity, queueType)
  }
}
