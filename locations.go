package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	esi "github.com/w9jds/go.esi"
)

var (
	characters = map[string]Character{}
)

func start() {
	go updateCharacters()
	go queueProcessor()
}

func queueProcessor() {
	for {
		select {
		case job := <-jobQueue:
			go processCharacter(job)
		}
	}
}

func processCharacter(job *Job) {
	time.Sleep(job.Delay)

	err := getCharacter(job.ID)
	if err != nil {
		log.Println(err)

		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	mutex.Lock()
	character := characters[job.ID]
	mutex.Unlock()

	if !hasLocationScopes(character.SSO) {
		jobQueue <- NewJob(job.ID, 60*time.Second)
		return
	}

	if ok, err := isCharacterOnline(character); !ok {
		if err != nil {
			log.Println(err)
		}

		jobQueue <- NewJob(job.ID, 60*time.Second)
		return
	}

	ship, location, err := getCharacterLocation(character)
	if err != nil {
		log.Println(err)
		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	names, err := esiClient.GetNames([]uint{ship.ShipTypeID, location.SolarSystemID})
	if err != nil {
		log.Println(err)
		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	err = pushLocation(character, ship, location, names)
	if err != nil {
		log.Println(err)

		jobQueue <- NewJob(job.ID, 15*time.Second)
		return
	}

	jobQueue <- NewJob(job.ID, 5*time.Second)
}

func isCharacterOnline(character Character) (bool, error) {
	online, err := esiClient.IsCharacterOnline(character.ID, character.SSO.AccessToken)
	if err != nil || !online.Online {
		database.NewRef(fmt.Sprintf("locations/%d", character.ID)).Delete(ctx)
		return false, err
	}

	return true, nil
}

func hasLocationScopes(permissions *Permissions) bool {
	if permissions == nil || permissions.Scope == "" {
		return false
	}

	if !strings.Contains(permissions.Scope, "read_location") {
		return false
	}
	if !strings.Contains(permissions.Scope, "read_ship_type") {
		return false
	}
	if !strings.Contains(permissions.Scope, "read_online") {
		return false
	}

	return true
}

func getCharacterLocation(character Character) (*esi.Ship, *esi.Location, error) {
	location, err := esiClient.GetCharacterLocation(character.ID, character.SSO.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	ship, err := esiClient.GetCharacterShip(character.ID, character.SSO.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	return ship, location, nil
}
