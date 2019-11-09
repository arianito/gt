package sms

import (
	"errors"
	"time"
)

type Mutex struct {
	lock       bool
	retries    int
	retryEvery int
}

func NewMutex(wait int, parts... int) *Mutex {
	part := 10
	if len(parts) > 0 {
		part = parts[0]
	}
	return &Mutex{
		lock:       false,
		retries:    part,
		retryEvery: wait / part,
	}
}

func (mutex *Mutex) Lock() (func(), error) {
	if mutex.lock {
		ch := make(chan bool)
		go func() {
			for i := 0; i < mutex.retries; i++ {
				time.Sleep(time.Millisecond * time.Duration(mutex.retryEvery))
				if !mutex.lock {
					ch <- true
					return
				}
			}
			ch <- false
		}()
		data := <-ch
		if !data {
			return nil, errors.New("failed to acquire lock")
		}
	}
	mutex.lock = true
	return func() {
		mutex.lock = false
	}, nil
}
