package main

import (
	"sync"
	"time"
)

var (
	jobQueue = make(chan *Job)

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
