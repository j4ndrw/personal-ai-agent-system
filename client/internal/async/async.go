package async

import "sync"

type Async struct {
	Coroutine func(done chan bool)
	Started   bool
	Done      chan bool
}

type Coroutine string
type table map[Coroutine]Async
type AsyncMap struct {
	table
	sync.Mutex
}

var Coroutines AsyncMap = AsyncMap{table: map[Coroutine]Async{}, Mutex: sync.Mutex{}}

func (am *AsyncMap) GetAsync(id Coroutine) (Async, bool) {
	am.Lock()
	a, ok := am.table[id]
	am.Unlock()
	return a, ok
}

func (am *AsyncMap) SetAsync(id Coroutine, a Async) {
	am.Lock()
	am.table[id] = a
	am.Unlock()
}


func (am *AsyncMap) RemoveAsync(id Coroutine) {
	am.Lock()
	delete(am.table, id)
	am.Unlock()
}

func Create(id string, coroutine func(done chan bool)) Coroutine {
	_id := Coroutine(id)
	if _, ok := Coroutines.GetAsync(_id); ok {
		return _id
	}

	a := Async{
		Coroutine: coroutine,
		Started:   false,
		Done:      make(chan bool),
	}
	Coroutines.SetAsync(_id, a)
	return _id
}

func (id Coroutine) Call() bool {
	a, ok := Coroutines.GetAsync(id)
	if !ok {
		return false
	}
	if a.Started {
		return false
	}

	a.Started = true
	Coroutines.SetAsync(id, a)
	go a.Coroutine(a.Done)
	return true
}

func (id Coroutine) Remove() {
	Coroutines.RemoveAsync(id)
}

func (id Coroutine) CallOnce() {
	if !id.Call() {
		return
	}

	go func() {
		for {
			coroutine, ok := Coroutines.GetAsync(id)
			if !ok { return }
			select {
			case <-coroutine.Done:
				id.Remove()
			}
		}
	}()
}
