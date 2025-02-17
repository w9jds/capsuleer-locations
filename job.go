package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func updateCharacter(key string) {
	item, err := rdb.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Error receiving character: %v", err)
		return
	}

	character := Character{}
	err = json.Unmarshal([]byte(item), &character)
	if err != nil {
		log.Printf("Error unmarshalling character: %v", err)
		return
	}

	mutex.Lock()
	characters[fmt.Sprintf("%d", character.ID)] = character
	mutex.Unlock()
}

func removeCharacter(id string) {
	mutex.Lock()
	delete(characters, id)
	mutex.Unlock()
}
