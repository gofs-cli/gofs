package repo

import (
	"log"
	"sync"
)

type Queue struct {
	Items []Item
	mutex sync.Mutex
}

type Item struct {
	Id     string
	Action string
}

func (q *Queue) IsBlocked(item Item) bool {
	if item.Action != "diagnostic" {
		return false
	}
	for _, qItem := range q.Items {
		if qItem.Action == "edit" {
			return true
		} else if qItem.Id == item.Id {
			return false
		}
	}
	return false
}

func (q *Queue) AddToQueue(item Item) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.Items = append(q.Items, item)
	log.Println(q)
}

func (q *Queue) RemoveFromQueue(item Item) {
	log.Println("removing item", item)
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i, qItem := range q.Items {
		if qItem.Id == item.Id {
			q.Items = append(q.Items[:i], q.Items[i+1:]...)
			log.Println(q)
			return
		}
	}
}
