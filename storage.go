package main

import (
	"fmt"
	"log"
	"time"

	esi "github.com/w9jds/go.esi"
)

type Character struct {
	ID         uint32       `json:"id,omitempty"`
	Name       string       `json:"name,omitempty"`
	AccountID  string       `json:"accountId,omitempty"`
	AllianceID uint32       `json:"allianceId,omitempty"`
	CorpID     uint32       `json:"corpId,omitempty"`
	Hash       string       `json:"hash,omitempty"`
	SSO        *Permissions `json:"sso,omitempty"`
	MemberFor  uint         `json:"memberFor,omitempty"`
	ETag       string
}

type Permissions struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
}

func getNowMilli() int64 {
	return (time.Now().UnixNano() / 1000000) - 30000
}

func updateCharacters() {
	lastUpdated := getNowMilli()
	var ids map[string]interface{}
	database.NewRef("characters").GetShallow(ctx, &ids)

	for id := range ids {
		if _, ok := characters[id]; !ok {
			jobQueue <- NewJob(id, 0*time.Second)
		}
	}

	for {
		time.Sleep(5 * time.Minute)
		var newlyAdded map[string]Character
		database.NewRef("characters").OrderByChild("createdAt").StartAt(lastUpdated).Get(ctx, &newlyAdded)
		lastUpdated = getNowMilli()

		for id := range newlyAdded {
			if _, ok := characters[id]; !ok {
				log.Println(fmt.Sprintf("new character %s found", id))
				jobQueue <- NewJob(id, 0*time.Second)
			}
		}
	}
}

func checkCharacterChanges(id string) error {
	var character Character
	ref := database.NewRef(fmt.Sprintf("characters/%s", id))

	if current, ok := getCharacter(id); ok {
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

func pushLocation(character Character, ship *esi.Ship, location *esi.Location, names map[uint]esi.NameRef) error {
	update := map[string]interface{}{
		"id":     character.ID,
		"name":   character.Name,
		"corpId": character.CorpID,
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
