package main

import (
	"fmt"
	"log"
	"time"

	esi "github.com/w9jds/go.esi"
)

type Ship struct {
	TypeID uint   `json:"typeId,omitempty"`
	Name   string `json:"name,omitempty"`
	ItemID uint   `json:"itemId,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Location struct {
	System *System `json:"system,omitempty"`
}

type System struct {
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func getNowMilli() int64 {
	return (time.Now().UnixNano() / 1000000) - 30000
}

func updateCharacters() {
	lastUpdated := getNowMilli()
	var ids map[string]interface{}
	var error = database.NewRef("characters").GetShallow(ctx, &ids)
	if error != nil {
		log.Fatalf("Error receiving character ids: %v", error)
	}

	for id := range ids {
		if _, ok := getCharacter(id); !ok {
			jobQueue <- NewJob(id, 0*time.Second)
		}
	}

	for {
		time.Sleep(5 * time.Minute)
		var newlyAdded map[string]Character
		database.NewRef("characters").OrderByChild("createdAt").StartAt(lastUpdated).Get(ctx, &newlyAdded)
		lastUpdated = getNowMilli()

		for id := range newlyAdded {
			if _, ok := getCharacter(id); !ok {
				log.Printf("new character %s found", id)
				jobQueue <- NewJob(id, 0*time.Second)
			}
		}
	}
}

func checkCharacterChanges(id string) error {
	var character Character
	ref := database.NewRef(fmt.Sprintf("characters/%s", id))

	if current, ok := getCharacter(id); ok {
		if current.SSO != nil {
			if expired, err := time.Parse(time.RFC3339Nano, current.SSO.ExpiresAt); err == nil {
				now := time.Now()
				diff := expired.Sub(now)

				if diff.Minutes() > 3 {
					return nil
				}
			}
		}

		hasChanged, etag, err := ref.GetIfChanged(ctx, current.ETag, &character)
		if err != nil {
			return err
		}

		if hasChanged {
			character.ETag = etag
			setCharacter(id, character)
		}
	} else {
		etag, err := ref.GetWithETag(ctx, &character)
		if err != nil {
			return err
		}

		character.ETag = etag
		setCharacter(id, character)
	}

	return nil
}

func pushLocation(character Character, ship *esi.Ship, location *esi.Location, names map[uint]esi.NameRef) error {
	update := map[string]interface{}{
		"accountId": character.AccountID,
		"id":        character.ID,
		"name":      character.Name,
		"corpId":    character.CorpID,
		"ship": &Ship{
			TypeID: ship.ShipTypeID,
			Name:   ship.ShipName,
			ItemID: ship.ShipItemID,
			Type:   names[ship.ShipTypeID].Name,
		},
		"location": &Location{
			System: &System{
				ID:   location.SolarSystemID,
				Name: names[location.SolarSystemID].Name,
			},
		},
	}

	if character.AllianceID != 0 {
		update["allianceId"] = character.AllianceID
	} else {
		update["allianceId"] = nil
	}

	err := database.NewRef(fmt.Sprintf("locations/%d", character.ID)).Update(ctx, update)
	if err != nil {
		return err
	}

	return nil
}
