package invoker

import (
	"database/sql"
	"encoding/json"

	"github.com/go-gorp/gorp"

	"github.com/ovh/cds/engine/api/application"
	"github.com/ovh/cds/engine/api/environment"
	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/sessionstore"
	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
)

// Insert inserts a new event listener in database after a new UUID has been generated
func Insert(db gorp.SqlExecutor, e *sdk.PipelineInvoker) error {
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
	if e.TypeID == 0 {
		e.TypeID = e.Type.ID
	}

	var edb PipelineInvoker
	edb = PipelineInvoker(*e)

	if err := db.Insert(&edb); err != nil {
		return err
	}

	return nil
}

//Delete deletes an event listener according to its uuid
func Delete(db gorp.SqlExecutor, uuid string) error {
	var edb PipelineInvoker
	edb = PipelineInvoker{UUID: uuid}

	n, err := db.Delete(&edb)
	if err != nil {
		return err
	}
	if n == 0 {
		return sdk.ErrPipelineInvokerNotFound
	}
	return nil
}

//Load loads an event listener according to its uuid
func Load(db gorp.SqlExecutor, uuid string) (*sdk.PipelineInvoker, error) {
	edb := PipelineInvoker{}
	if err := db.SelectOne(&edb, "select * from pipeline_invoker where uuid = $1", uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, sdk.ErrPipelineInvokerNotFound
		}
		return nil, err
	}
	e := sdk.PipelineInvoker(edb)
	return &e, nil
}

//PostGet is a DB hook
func (e *PipelineInvoker) PostGet(db gorp.SqlExecutor) error {
	app, err := application.LoadByID(db, e.ApplicationID, nil)
	if err != nil {
		log.Warning("source.PostGest> Unable load application %d: %s", e.ApplicationID, err)
		return err
	}
	pip, err := pipeline.LoadPipelineByID(db, e.PipelineID, true)
	if err != nil {
		log.Warning("source.PostGest> Unable load pipeline %d: %s", e.PipelineID, err)
		return err
	}
	env, err := environment.LoadEnvironmentByID(db, e.EnvironmentID)
	if err != nil {
		log.Warning("source.PostGest> Unable load env %d: %s", e.EnvironmentID, err)
		return err
	}
	t, err := LoadInvokerType(db, e.TypeID)
	if err != nil {
		log.Warning("source.PostGest> Unable load pipeline invoker type %d: %s", e.TypeID, err)
		return err
	}
	e.Application = *app
	e.Pipeline = *pip
	e.Environment = *env
	e.Type = *t
	return nil
}

// InsertInvokerType inserts a PipelineInvokerType in DB
func InsertInvokerType(db gorp.SqlExecutor, e *sdk.PipelineInvokerType) error {
	t := PipelineInvokerType(*e)
	if err := db.Insert(&t); err != nil {
		return err
	}
	*e = (sdk.PipelineInvokerType(t))
	return nil
}

// UpdateInvokerType updates a PipelineInvokerType in DB
func UpdateInvokerType(db gorp.SqlExecutor, e *sdk.PipelineInvokerType) error {
	t := PipelineInvokerType(*e)
	n, err := db.Update(&t)
	if err != nil {
		log.Warning("UpdateInvokerType> Unable to update %d :%s", e.ID, err)
		return err
	}
	if n == 0 {
		log.Warning("UpdateInvokerType> Unable to update %d; not found", e.ID)
		return sdk.ErrPipelineInvokerNotFound
	}
	*e = (sdk.PipelineInvokerType(t))
	return nil
}

// DeleteInvokerType deletes an event listener according to its uuid
func DeleteInvokerType(db gorp.SqlExecutor, id int64) error {
	var edb PipelineInvokerType
	edb = PipelineInvokerType{ID: id}

	n, err := db.Delete(&edb)
	if err != nil {
		return err
	}
	if n == 0 {
		return sdk.ErrPipelineInvokerNotFound
	}
	return nil
}

// LoadInvokerType loads a pipeline invoker type given its id
func LoadInvokerType(db gorp.SqlExecutor, id int64) (*sdk.PipelineInvokerType, error) {
	var edb PipelineInvokerType
	if err := db.SelectOne(&edb, "select * from pipeline_invoker_type where id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, sdk.ErrPipelineInvokerNotFound
		}
		return nil, err
	}
	e := sdk.PipelineInvokerType(edb)
	return &e, nil
}

// LoadAllInvokerTypes loads all pipeline invoker types
func LoadAllInvokerTypes(db gorp.SqlExecutor) ([]sdk.PipelineInvokerType, error) {
	var tdbs []PipelineInvokerType
	if _, err := db.Select(&tdbs, "select * from pipeline_invoker_type"); err != nil {
		return nil, err
	}
	t := []sdk.PipelineInvokerType{}
	for _, p := range tdbs {
		if err := p.PostGet(db); err != nil {
			return nil, err
		}
		t = append(t, sdk.PipelineInvokerType(p))
	}
	return t, nil
}

// PostGet is a DB hook
func (e *PipelineInvokerType) PostGet(db gorp.SqlExecutor) error {
	str, err := db.SelectStr("select type from pipeline_invoker_type where id = $1", e.ID)
	if err != nil {
		return err
	}

	dockerType := &sdk.PipelineInvokerTypeDocker{}
	localType := &sdk.PipelineInvokerTypeLocal{}

	if err := json.Unmarshal([]byte(str), &dockerType); err != nil {
		return err
	}

	if dockerType.DockerImage != "" {
		e.DockerType = dockerType
		return nil
	}

	if err := json.Unmarshal([]byte(str), &localType); err != nil {
		return err
	}

	e.LocalType = localType
	return nil
}

// PostInsert is a DB hook
func (e *PipelineInvokerType) PostInsert(db gorp.SqlExecutor) error {
	query := "update pipeline_invoker_type set type = $2 where id = $1"
	var btes []byte
	var err error
	if e.DockerType != nil {
		btes, err = json.Marshal(e.DockerType)
		if err != nil {
			return err
		}
	} else {
		btes, err = json.Marshal(e.LocalType)
		if err != nil {
			return err
		}
	}
	_, err = db.Exec(query, e.ID, btes)
	return err
}

// PostUpdate is a DB hook
func (e *PipelineInvokerType) PostUpdate(db gorp.SqlExecutor) error {
	return e.PostInsert(db)
}
