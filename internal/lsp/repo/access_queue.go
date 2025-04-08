package repo

import (
	"sync"
)

// AccessQueue tracks the order of requests to read/update a repo in order to prevent
// diagnostics from accessing data until the related edits have been performed.
type AccessQueue struct {
	Items []Item
	mutex sync.Mutex
}

type Item struct {
	Id string
}

func (a *AccessQueue) AwaitUnblock(item Item) {
	for {
		if len(a.Items) == 0 || a.Items[0].Id == item.Id {
			return
		}
	}
}

func (a *AccessQueue) Add(item Item) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Items = append(a.Items, item)
}

func (a *AccessQueue) Remove(item Item) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for i, queueItem := range a.Items {
		if queueItem.Id == item.Id {
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return
		}
	}
}
