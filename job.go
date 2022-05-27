package main

import (
	"log"
	"strconv"
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
	// TODO Find out why I'm not getting an ID from some characters
	if value.ID < 1 {
		id, err := strconv.Atoi(id)
		if err != nil {
			log.Println(err)
		}

		value.ID = uint32(id)
	}

	mutex.Lock()
	characters[id] = value
	mutex.Unlock()
}
