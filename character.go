package main

import (
	"fmt"
	"log"
	"strings"
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

func (character *Character) isCharacterOnline() (bool, error) {
	online, err := esiClient.IsCharacterOnline(character.ID, character.SSO.AccessToken)
	if err != nil || !online.Online {
		database.NewRef(fmt.Sprintf("locations/%d", character.ID)).Delete(ctx)
		return false, err
	}

	return true, nil
}

func (character *Character) getCharacterLocation() (*esi.Ship, *esi.Location, error) {
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

func (permissions *Permissions) isAuthenticated(id string) bool {
	if permissions == nil || permissions.AccessToken == "" {
		return false
	}

	now := time.Now()
	expired, err := time.Parse(time.RFC3339Nano, permissions.ExpiresAt)
	if err != nil {
		log.Println(err)
		return false
	}

	isValid := now.Before(expired)

	if !isValid {
		diff := now.Sub(expired)
		if diff.Hours() > 1 {
			log.Printf("%s token has expired", id)
			database.NewRef(fmt.Sprintf("locations/%s", id)).Delete(ctx)
			database.NewRef(fmt.Sprintf("characters/%s/sso", id)).Delete(ctx)
			database.NewRef(fmt.Sprintf("characters/%s/titles", id)).Delete(ctx)
			database.NewRef(fmt.Sprintf("characters/%s/roles", id)).Delete(ctx)
		}
	}

	return isValid
}

func (permissions *Permissions) hasLocationScopes() bool {
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
