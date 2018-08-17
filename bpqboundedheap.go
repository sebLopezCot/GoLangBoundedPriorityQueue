package bpq

import (
  "encoding/json"
  "fmt"
  "math"
)

type entry struct {
  item     interface{}
  priority float64
  inUse    bool
}

type bpqBoundedHeapImpl struct {
  entries              []entry
  nextSlot             int
  highestPriorityIndex int
  compareFunc          func(float64, float64) bool
  queueType            QueueType
}

func (bpq bpqBoundedHeapImpl) String() string {
  return fmt.Sprintf("BPQ entries: %v", bpq.entries)
}

func (e entry) String() string {
  return fmt.Sprintf("{ Value %v, priority %v }", e.item, e.priority)
}

func makeBoundedHeap(capacity int, queueType QueueType) *bpqBoundedHeapImpl {

  var fn func(float64, float64) bool
  if queueType == MaxQueue {
    fn = func(a float64, b float64) bool {
      return a > b
    }
  } else {
    fn = func(a float64, b float64) bool {
      return a < b
    }
  }

  result := bpqBoundedHeapImpl{make([]entry, capacity), 0, 0, fn, queueType}

  for i := 0; i < capacity; i++ {
    result.entries[i] = entry{nil, 0, false}
  }

  return &result
}

func (bpq *bpqBoundedHeapImpl) Capacity() int {
  return len(bpq.entries)
}

func (bpq *bpqBoundedHeapImpl) QueueType() QueueType {
  return bpq.queueType
}

func (bpq *bpqBoundedHeapImpl) Push(item interface{}, priority float64) bool {
  if bpq.Capacity() == bpq.nextSlot {
    // We're full!
    if bpq.compareFunc(priority, bpq.entries[bpq.highestPriorityIndex].priority) {
      // But we can knock the back entry off
      bpq.entries[bpq.highestPriorityIndex].item = item
      bpq.entries[bpq.highestPriorityIndex].priority = priority
      bpq.entries[bpq.highestPriorityIndex].inUse = true

      bpq.bubbleUpIndex(bpq.highestPriorityIndex)

      // We must now restore the highest priority index. Alas, this is an O(n)
      // operation, but it only triggers when insertion is successful
      // and the heap is full
      highestPriority := float64(math.MinInt64)
      highestIndex := 0
      for i, v := range bpq.entries {
        if bpq.compareFunc(highestPriority, v.priority) {
          highestIndex = i
          highestPriority = v.priority
        }
      }
      bpq.highestPriorityIndex = highestIndex

      return true
    } else {
      return false
    }
  } else {
    // We're not full; add in the result
    bpq.entries[bpq.nextSlot].item = item
    bpq.entries[bpq.nextSlot].priority = priority
    bpq.entries[bpq.nextSlot].inUse = true

    // If we're highest than the highest, we become the highest
    // We've guarnteed to not bubble up in such cases
    if bpq.compareFunc(bpq.entries[bpq.highestPriorityIndex].priority, priority) {
      bpq.highestPriorityIndex = bpq.nextSlot
    } else {
      bpq.bubbleUpIndex(bpq.nextSlot)
    }

    bpq.nextSlot = bpq.nextSlot + 1

    return true
  }
}

func (bpq *bpqBoundedHeapImpl) Pop() (QueueItem, error) {
  // defer fmt.Printf("Post-pop %v\n", bpq)

  if bpq.nextSlot == 0 {
    return QueueItem{}, NoElementsError
  }

  resultEntry := bpq.entries[0]

  // Are we moving the highest priority index
  poppingHighest := bpq.nextSlot-1 == bpq.highestPriorityIndex

  bpq.entries[0] = bpq.entries[bpq.nextSlot-1]
  bpq.entries[bpq.nextSlot-1].inUse = false
  bpq.nextSlot = bpq.nextSlot - 1

  if bpq.entries[0].inUse {
    newIndex := bpq.bubbleDownIndex(0)

    if poppingHighest {
      bpq.highestPriorityIndex = newIndex
    }
  }

  result := QueueItem{
    Value:    resultEntry.item,
    Priority: resultEntry.priority,
  }

  return result, nil
}

func (bpq *bpqBoundedHeapImpl) MarshalJSON() ([]byte, error) {
  buffer := make([]QueueItem, 0, 0)
  for {
    val, err := bpq.Pop()
    if err != nil {
      break
    }

    buffer = append(buffer, val)
  }

  return json.Marshal(buffer)
}

func (bpq *bpqBoundedHeapImpl) bubbleDownIndex(index int) int {
  // If we're greater than either of our children, swap with the child
  // and bubble down from their
  lChild, rChild := childrenOfIndex(index)

  if lChild < bpq.nextSlot && rChild < bpq.nextSlot {
    if bpq.compareFunc(bpq.entries[lChild].priority, bpq.entries[rChild].priority) {
      if bpq.compareFunc(bpq.entries[lChild].priority, bpq.entries[index].priority) {
        bpq.swapIndices(lChild, index)
        return bpq.bubbleDownIndex(lChild)
      }
    } else {
      if bpq.compareFunc(bpq.entries[rChild].priority, bpq.entries[index].priority) {
        bpq.swapIndices(rChild, index)
        return bpq.bubbleDownIndex(rChild)
      }
    }
  } else if lChild < bpq.nextSlot && bpq.compareFunc(bpq.entries[lChild].priority, bpq.entries[index].priority) {
    bpq.swapIndices(lChild, index)
    return bpq.bubbleDownIndex(lChild)
  } else if rChild < bpq.nextSlot && bpq.compareFunc(bpq.entries[rChild].priority, bpq.entries[index].priority) {
    bpq.swapIndices(rChild, index)
    return bpq.bubbleDownIndex(rChild)
  }

  return index
}

func (bpq *bpqBoundedHeapImpl) bubbleUpIndex(index int) {
  parent := parentOfIndex(index)

  if bpq.compareFunc(bpq.entries[index].priority, bpq.entries[parent].priority) {
    bpq.swapIndices(index, parent)
    bpq.bubbleUpIndex(parent)
  }
}

func (bpq *bpqBoundedHeapImpl) swapIndices(left, right int) {
  bpq.entries[left], bpq.entries[right] = bpq.entries[right], bpq.entries[left]
}

func parentOfIndex(index int) int {
  return (index - 1) / 2
}

func childrenOfIndex(index int) (int, int) {
  return (index * 2) + 1, (index * 2) + 2
}
