/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"runtime"

	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	"github.com/r3labs/natsdb"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var n *nats.Conn
var db *gorm.DB
var err error
var handler natsdb.Handler

func startHandler() {
	handler = natsdb.Handler{
		NotFoundErrorMessage:   natsdb.NotFound.Encoded(),
		UnexpectedErrorMessage: natsdb.Unexpected.Encoded(),
		DeletedMessage:         []byte(`{"status":"deleted"}`),
		Nats:                   n,
		NewModel: func() natsdb.Model {
			return &Entity{}
		},
	}

	if _, err = n.Subscribe("authorization.get", handler.Get); err != nil {
		log.Println("Error subscribing authorization.get")
	}
	if _, err = n.Subscribe("authorization.del", handler.Del); err != nil {
		log.Println("Error subscribing authorization.del")
	}
	if _, err = n.Subscribe("authorization.set", handler.Set); err != nil {
		log.Println("Error subscribing authorization.set")
	}
	if _, err = n.Subscribe("authorization.find", handler.Find); err != nil {
		log.Println("Error subscribing authorization.find")
	}

}

func main() {
	setupNats()
	setupPg()
	startHandler()

	runtime.Goexit()
}
