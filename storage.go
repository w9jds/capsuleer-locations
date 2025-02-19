package main

import (
	"fmt"
	"log"
	"strings"
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

func updateCharacters() {
	characters, err := rdb.Keys(ctx, "characters:*").Result()
	if err != nil {
		log.Fatalf("Error receiving character ids: %v", err)
	}

	for _, key := range characters {
		updateCharacter(key)
		path := strings.Split(key, ":")

		jobQueue <- NewJob(path[1], 0*time.Second)
	}

	pubsub := rdb.PSubscribe(ctx, "__keyspace@*__:characters:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		switch msg.Payload {
		case "set":
			path := strings.Split(msg.Channel, ":")
			_, ok := getCharacter(path[2])

			updateCharacter(fmt.Sprintf("%s:%s", path[1], path[2]))

			if !ok {
				// new character found, add to queue
				jobQueue <- NewJob(path[2], 0*time.Second)
			}
		case "del":
			path := strings.Split(msg.Channel, ":")
			removeCharacter(path[2])
		}
	}
}

func getNames(systemId uint, shipId uint) (map[uint]string, error) {
	missing := []uint{}
	names := make(map[uint]string)

	systemName, err := rdb.HGet(ctx, "names:solar_system", fmt.Sprint(systemId)).Result()
	if err == nil {
		names[systemId] = systemName
	} else {
		missing = append(missing, systemId)
	}

	shipName, err := rdb.HGet(ctx, "names:inventory_type", fmt.Sprint(systemId)).Result()
	if err == nil {
		names[shipId] = shipName
	} else {
		missing = append(missing, shipId)
	}

	if len(missing) > 0 {
		refs, err := esiClient.GetNames(missing)
		if err != nil {
			return nil, err
		}

		cacheNames(refs)
		for id, ref := range refs {
			names[id] = ref.Name
		}
	}

	return names, nil
}

func cacheNames(names map[uint]esi.NameRef) {
	for id, name := range names {
		rdb.HSet(ctx, "names", id, name.Name)
	}
}

func pushLocation(character Character, ship *esi.Ship, location *esi.Location, names map[uint]string) error {
	update := map[string]interface{}{
		"accountId": character.AccountID,
		"id":        character.ID,
		"name":      character.Name,
		"corpId":    character.CorpID,
		"ship": &Ship{
			TypeID: ship.ShipTypeID,
			Name:   ship.ShipName,
			ItemID: ship.ShipItemID,
			Type:   names[ship.ShipTypeID],
		},
		"location": &Location{
			System: &System{
				ID:   location.SolarSystemID,
				Name: names[location.SolarSystemID],
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
