/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// Entity : the database mapped entity
type Entity struct {
	ID           uint   `json:"id" gorm:"primary_key"`
	UserID       string `json:"user_id" gorm:"unique_index:idx_uniq"`
	ResourceID   string `json:"resource_id" gorm:"unique_index:idx_uniq"`
	ResourceType string `json:"resource_type" gorm:"unique_index:idx_uniq"`
	Role         string `json:"role"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be authorization
func (Entity) TableName() string {
	return "authorizations"
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (e *Entity) Find() (list []interface{}) {
	entities := []Entity{}
	if e.UserID == "" {
		db.Where("resource_id = ? AND resource_type = ?", e.ResourceID, e.ResourceType).Find(&entities)
	} else if e.ResourceID == "" {
		db.Where("user_id = ? AND resource_type = ?", e.UserID, e.ResourceType).Find(&entities)
	} else if e.Role == "" {
		db.Where("user_id = ? AND resource_id = ? AND resource_type = ?", e.UserID, e.ResourceID, e.ResourceType).Find(&entities)
	} else if e.Role == "" && e.ResourceID == "" && e.UserID == "" {
		db.Find(&entities)
	} else {
		db.Where("user_id = ? AND resource_id = ? AND resource_type = ? AND role = ?", e.UserID, e.ResourceID, e.ResourceType, e.Role).Find(&entities)
	}
	list = make([]interface{}, len(entities))
	for i, s := range entities {
		list[i] = s
	}

	return list
}

// MapInput : maps the input []byte on the current entity
func (e *Entity) MapInput(body []byte) {
	if err := json.Unmarshal(body, &e); err != nil {
		log.Println("Invalid input " + err.Error())
	}
}

// HasID : determines if the current entity has an id or not
func (e *Entity) HasID() bool {
	if e.ID == 0 {
		return false
	}
	return true
}

// LoadFromInput : Will load from a []byte input the database stored entity
func (e *Entity) LoadFromInput(msg []byte) bool {
	e.MapInput(msg)
	var stored Entity
	if e.ID > 0 {
		db.Where("id = ?", e.ID).First(&stored)
	} else {
		db.Where("user_id = ? AND resource_id = ? AND resource_type = ? AND role = ?", e.UserID, e.ResourceID, e.ResourceType, e.Role).First(&stored)
	}

	if &stored == nil {
		return false
	}
	if ok := stored.HasID(); !ok {
		return false
	}

	e.ID = stored.ID
	e.UserID = stored.UserID
	e.ResourceID = stored.ResourceID
	e.ResourceType = stored.ResourceType
	e.Role = stored.Role
	e.CreatedAt = stored.CreatedAt
	e.UpdatedAt = stored.UpdatedAt

	return true
}

// LoadFromInputOrFail : Will try to load from the input an existing entity,
// or will call the handler to Fail the nats message
func (e *Entity) LoadFromInputOrFail(msg *nats.Msg, h *natsdb.Handler) bool {
	stored := &Entity{}
	ok := stored.LoadFromInput(msg.Data)
	if !ok {
		h.Fail(msg)
	}
	*e = *stored

	return ok
}

// Update : It will update the current entity with the input []byte
func (e *Entity) Update(body []byte) error {
	e.MapInput(body)
	stored := Entity{}
	if e.ID > 0 {
		db.Where("id = ?", e.ID).First(&stored)
	} else {
		db.Where("user_id = ? AND resource_id = ? AND resource_type = ?", e.UserID, e.ResourceID, e.ResourceType).First(&stored)
	}

	stored.Role = e.Role

	db.Save(&stored)
	e = &stored

	return nil
}

// Delete : Will delete from database the current Entity
func (e *Entity) Delete() error {
	db.Unscoped().Delete(&e)

	return nil
}

// Save : Persists current entity on database
func (e *Entity) Save() error {
	db.Save(&e)

	return nil
}
