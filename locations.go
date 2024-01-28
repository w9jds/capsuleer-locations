package main

import (
	"log"
	"time"
)

func start() {
	go updateCharacters()
	go queueProcessor()
}

func queueProcessor() {
	for job := range jobQueue {
		go processCharacter(job)
	}
}

func processCharacter(job *Job) {
	time.Sleep(job.Delay)

	err := checkCharacterChanges(job.ID)
	if err != nil {
		log.Printf("Character %s: %s", job.ID, err)
		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	character, _ := getCharacter(job.ID)

	if !character.SSO.isAuthenticated(job.ID) {
		jobQueue <- NewJob(job.ID, 10*time.Minute)
		return
	}

	if !character.SSO.hasLocationScopes() {
		jobQueue <- NewJob(job.ID, 10*time.Minute)
		return
	}

	if ok, err := character.isCharacterOnline(); !ok {
		if err != nil {
			log.Printf("Error receiving character online: %v", err)
		}

		jobQueue <- NewJob(job.ID, 60*time.Second)
		return
	}

	ship, location, err := character.getCharacterLocation()
	if err != nil {
		log.Printf("Error receiving character location/ship: %v", err)
		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	names, err := esiClient.GetNames([]uint{ship.ShipTypeID, location.SolarSystemID})
	if err != nil {
		log.Printf("Error receiving names for [%v, %v]: %v", ship.ShipTypeID, location.SolarSystemID, err)
		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	err = pushLocation(character, ship, location, names)
	if err != nil {
		log.Printf("Error updating characters location: %v", err)

		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	jobQueue <- NewJob(job.ID, 5*time.Second)
}
