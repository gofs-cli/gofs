package repo

import (
	"sync"
)

const (
	ACTION_DIAGNOSTIC = "diagnostic"
	ACTION_EDIT       = "edit"
)

// AccessQueue tracks handlers wishing to read/update a repo in order to prevent
// diagnostics from accessing data until the related edits have been performed.
type AccessQueue struct {
	Items []Item
	mutex sync.Mutex
}

type Item struct {
	Id     string
	Action string
}

func (a *AccessQueue) AwaitUnblock(item Item) {
	if item.Action != ACTION_DIAGNOSTIC {
		return
	}
	for {
		for _, queueItem := range a.Items {
			if queueItem.Action == ACTION_EDIT {
				break
			} else if queueItem.Id == item.Id {
				return
			}
		}
	}
}

func (a *AccessQueue) AddToQueue(item Item) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.Items = append(a.Items, item)
}

func (a *AccessQueue) RemoveFromQueue(item Item) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for i, queueItem := range a.Items {
		if queueItem.Id == item.Id {
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return
		}
	}
}
