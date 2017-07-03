package structs

type Queue struct {
	items []interface{}
}

func (q *Queue) Length() int {
	return len(q.items)
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() interface{} {
	if q.IsEmpty() {
		return ""
	}
	val := q.items[0]
	q.items = q.items[1:len(q.items)]
	return val
}
