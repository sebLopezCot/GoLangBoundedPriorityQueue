package bpq

import (
  "encoding/json"
  "errors"
  "fmt"
)

type QueueType int

type QueueItem struct {
  Value    interface{}
  Priority float64
}

type QueueMarshalled struct {
  Items    []QueueItem
  Capacity int
  Ordering QueueType
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

func (queueType QueueType) MarshalJSON() ([]byte, error) {
  return json.Marshal(queueType.String())
}

func (queueType *QueueType) UnmarshalJSON(text []byte) error {
  textStr := string(text)

  for i, typeString := range queueTypes {
    if "\""+typeString+"\"" == textStr {
      *queueType = QueueType(i + 1)
      return nil
    }
  }

  return errors.New("could not find corresponding type for unmarshalling string")
}

// NoElementsError indicates that a Pop was attempted on an empty queue
var NoElementsError error = errors.New("NoElementsError")

const maxRingBufferSize = 128

// BPQWithCapacity creates a new bounded priority queue with the given capacity
func BPQWithCapacity(capacity int, queueType QueueType) *BPQ {
  return makeRingBuffer(capacity, queueType)
}

type ringBufferEntry struct {
  value    interface{}
  priority float64
  inUse    bool
}

type BPQ struct {
  entries              []ringBufferEntry
  startIndex, endIndex int
  compareFunc          func(float64, float64) bool
  queueType            QueueType
}

func (e ringBufferEntry) String() string {
  if e.inUse {
    return fmt.Sprintf("{ Value %v, Priority %v }", e.value, e.priority)
  } else {
    return fmt.Sprintf("{}")
  }
}

func (bpq *BPQ) String() string {
  return fmt.Sprintf("BPQ Ring Buffer: Start %v, End %v, Entries: %v", bpq.startIndex, bpq.endIndex, bpq.entries)
}

func makeRingBuffer(capacity int, queueType QueueType) *BPQ {
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

  result := BPQ{make([]ringBufferEntry, capacity),
    0, 0, fn, queueType}

  for i := 0; i < capacity; i++ {
    result.entries[i] = ringBufferEntry{nil, 0, false}
  }

  return &result
}

func (bpq *BPQ) Capacity() int {
  return len(bpq.entries)
}

func (bpq *BPQ) Push(item interface{}, priority float64) bool {
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

func (bpq *BPQ) Pop() (QueueItem, error) {
  if bpq.entries[bpq.startIndex].inUse == false {
    return QueueItem{}, NoElementsError
  }

  result := QueueItem{
    Value:    bpq.entries[bpq.startIndex].value,
    Priority: bpq.entries[bpq.startIndex].priority,
  }
  bpq.entries[bpq.startIndex].inUse = false

  bpq.startIndex = bpq.startIndex + 1
  if bpq.startIndex == len(bpq.entries) {
    bpq.startIndex = 0
  }

  return result, nil
}

func (bpq *BPQ) MarshalJSON() ([]byte, error) {
  buffer := make([]QueueItem, 0, 0)
  for {
    item, err := bpq.Pop()
    if err != nil {
      break
    }

    buffer = append(buffer, item)
  }

  return json.Marshal(QueueMarshalled{
    Items:    buffer,
    Capacity: cap(bpq.entries),
    Ordering: bpq.queueType,
  })
}

func (bpq *BPQ) UnmarshalJSON(data []byte) error {
  var buffer QueueMarshalled
  if err := json.Unmarshal(data, &buffer); err != nil {
    return err
  }

  *bpq = *BPQWithCapacity(buffer.Capacity, buffer.Ordering)

  for _, item := range buffer.Items {
    bpq.Push(item.Value, item.Priority)
  }

  return nil
}
