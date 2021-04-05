package main

import (
	"fmt"
	"time"

	esi "github.com/w9jds/go.esi"
)

type Character struct {
	ID         uint32          `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`
	AccountID  string          `json:"accountId,omitempty"`
	AllianceID uint32          `json:"allianceId,omitempty"`
	CorpID     uint32          `json:"corpId,omitempty"`
	Titles     interface{}     `json:"titles,omitempty"`
	Hash       string          `json:"hash,omitempty"`
	Roles      *CharacterRoles `json:"roles,omitempty"`
	SSO        *Permissions    `json:"sso,omitempty"`
	MemberFor  uint            `json:"memberFor,omitempty"`
	ETag       string
}

type CharacterLocation struct {
	ID         uint32    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	CorpID     uint32    `json:"corpId,omitempty"`
	AllinaceID uint32    `json:"allianceId,omitempty"`
	Ship       *Ship     `json:"ship,omitempty"`
	Location   *Location `json:"location,omitempty"`
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

type CharacterRoles struct {
	Roles        interface{} `json:"roles,omitempty"`
	RolesAtHQ    interface{} `json:"roles_at_hq,omitempty"`
	RolesAtOther interface{} `json:"roles_at_other,omitempty"`
}

type Permissions struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
}

func updateCharacters() {
	for {
		var ids map[string]interface{}
		database.NewRef("characters").GetShallow(ctx, &ids)

		for id := range ids {
			if _, ok := characters[id]; !ok {
				jobQueue <- NewJob(id, 0*time.Second)
			}
		}

		time.Sleep(5 * time.Minute)
	}
}

func getCharacter(id string) error {
	var character Character
	ref := database.NewRef(fmt.Sprintf("characters/%s", id))

	mutex.Lock()
	current, ok := characters[id]
	mutex.Unlock()

	if ok {
		hasChanged, etag, err := ref.GetIfChanged(ctx, current.ETag, &character)
		if err != nil {
			return err
		}

		if hasChanged {
			character.ETag = etag
			mutex.Lock()
			characters[id] = character
			mutex.Unlock()
		}
	} else {
		etag, err := ref.GetWithETag(ctx, &character)
		if err != nil {
			return err
		}

		character.ETag = etag
		mutex.Lock()
		characters[id] = character
		mutex.Unlock()
	}

	return nil
}

func pushLocation(character Character, ship *esi.Ship, location *esi.Location, names map[uint]esi.NameRef) error {
	update := &CharacterLocation{
		ID:         character.ID,
		Name:       character.Name,
		CorpID:     character.CorpID,
		AllinaceID: character.AllianceID,
		Ship: &Ship{
			TypeID: ship.ShipTypeID,
			Name:   ship.ShipName,
			ItemID: ship.ShipItemID,
			Type:   names[ship.ShipTypeID].Name,
		},
		Location: &Location{
			System: &System{
				ID:   location.SolarSystemID,
				Name: names[location.SolarSystemID].Name,
			},
		},
	}

	err := database.NewRef(fmt.Sprintf("locations/%d", character.ID)).Set(ctx, update)
	if err != nil {
		return err
	}

	return nil
}
