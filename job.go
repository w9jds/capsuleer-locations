package main

import (
	"sync"
	"time"
)

var (
	characters = map[string]Character{}
	jobQueue   = make(chan *Job)

	mutex = &sync.Mutex{}
)

type Job struct {
	ID    string
	Delay time.Duration
}

func NewJob(id string, delay time.Duration) *Job {
	return &Job{
		ID:    id,
		Delay: delay,
	}
}

func getCharacter(id string) (Character, bool) {
	mutex.Lock()
	character, ok := characters[id]
	mutex.Unlock()

	return character, ok
}

func setCharacter(id string, value Character) {
	mutex.Lock()
	characters[id] = value
	mutex.Unlock()
}
