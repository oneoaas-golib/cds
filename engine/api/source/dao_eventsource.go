package source

import (
	"github.com/go-gorp/gorp"

	"database/sql"

	"github.com/ovh/cds/engine/api/application"
	"github.com/ovh/cds/engine/api/environment"
	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/sessionstore"
	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
)

// Insert inserts a new event listener in database after a new UUID has been generated
func Insert(db gorp.SqlExecutor, e *sdk.EventListener) error {
	uuid, err := sessionstore.NewSessionKey()
	if err != nil {
		return err
	}

	e.UUID = string(uuid)

	if e.ApplicationID == 0 {
		e.ApplicationID = e.Application.ID
	}
	if e.EnvironmentID == 0 {
		e.EnvironmentID = e.Environment.ID
	}
	if e.PipelineID == 0 {
		e.PipelineID = e.Pipeline.ID
	}

	var edb EventListener
	edb = EventListener(*e)

	if err := db.Insert(&edb); err != nil {
		return err
	}

	return nil
}

//Delete deletes an event listener according to its uuid
func Delete(db gorp.SqlExecutor, uuid string) error {
	var edb EventListener
	edb = EventListener{UUID: uuid}

	n, err := db.Delete(&edb)
	if err != nil {
		return err
	}
	if n == 0 {
		return sdk.ErrEventListenerNotFound
	}
	return nil
}

//Load loads an event listener according to its uuid
func Load(db gorp.SqlExecutor, uuid string) (*sdk.EventListener, error) {
	edb := EventListener{}
	if err := db.SelectOne(&edb, "select * from event_listener where uuid = $1", uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, sdk.ErrEventListenerNotFound
		}
		return nil, err
	}
	e := sdk.EventListener(edb)
	return &e, nil
}

//PostGet is a DB hook
func (e *EventListener) PostGet(db gorp.SqlExecutor) error {
	app, err := application.LoadApplicationByID(db, e.ApplicationID)
	if err != nil {
		log.Warning("source.PostGest> Unable load load application %d: %s", e.ApplicationID, err)
		return err
	}
	pip, err := pipeline.LoadPipelineByID(db, e.PipelineID, true)
	if err != nil {
		log.Warning("source.PostGest> Unable load load pipeline %d: %s", e.PipelineID, err)
		return err
	}
	env, err := environment.LoadEnvironmentByID(db, e.EnvironmentID)
	if err != nil {
		log.Warning("source.PostGest> Unable load load env %d: %s", e.EnvironmentID, err)
		return err
	}
	e.Application = *app
	e.Pipeline = *pip
	e.Environment = *env
	return nil
}
