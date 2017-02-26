package invoker

import (
	"testing"

	"github.com/fsamin/go-dump"
	"github.com/stretchr/testify/assert"

	"github.com/ovh/cds/engine/api/application"
	"github.com/ovh/cds/engine/api/bootstrap"
	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/engine/api/test/assets"
	"github.com/ovh/cds/sdk"
)

func TestInsertInvokerType(t *testing.T) {
	db := test.SetupPG(t)

	pit := sdk.PipelineInvokerType{
		Author:      "author",
		Description: "description",
		Identifier:  "identifier",
		Name:        "name",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img",
			DockerCmd:   "cmd",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	pit1, err := LoadInvokerType(db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(db, pit.ID))
}

func TestUpdateInvokerType(t *testing.T) {
	db := test.SetupPG(t)

	pit := sdk.PipelineInvokerType{
		Author:      "author",
		Description: "description",
		Identifier:  "identifier",
		Name:        "name",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img",
			DockerCmd:   "cmd",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	pit1, err := LoadInvokerType(db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)

	pit.Description = "description2"
	pit.LocalType = &sdk.PipelineInvokerTypeLocal{
		LocalCmd: "local",
	}
	pit.DockerType = nil

	test.NoError(t, UpdateInvokerType(db, &pit))
	pit1, err = LoadInvokerType(db, pit.ID)

	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(db, pit.ID))
}

func TestDeleteInvokerType(t *testing.T) {
	db := test.SetupPG(t)
	pit := sdk.PipelineInvokerType{
		Author:      "author",
		Description: "description",
		Identifier:  "identifier",
		Name:        "name",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img",
			DockerCmd:   "cmd",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	pit1, err := LoadInvokerType(db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(db, pit.ID))

	_, err = LoadInvokerType(db, pit.ID)
	assert.Equal(t, sdk.ErrPipelineInvokerNotFound, err)
}

func TestLoadInvokerType(t *testing.T) {
	//covered by TestInsertInvokerType
}

func TestLoadAllInvokerTypes(t *testing.T) {
	db := test.SetupPG(t)

	pit := sdk.PipelineInvokerType{
		Author:      "author",
		Description: "description",
		Identifier:  "identifier",
		Name:        "name",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img",
			DockerCmd:   "cmd",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	pit = sdk.PipelineInvokerType{
		Author:      "author2",
		Description: "description2",
		Identifier:  "identifier2",
		Name:        "name2",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img2",
			DockerCmd:   "cmd2",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	is, err := LoadAllInvokerTypes(db)
	test.NoError(t, err)
	assert.Equal(t, 2, len(is))

	for _, i := range is {
		test.NoError(t, DeleteInvokerType(db, i.ID))
	}
}

func TestInsert(t *testing.T) {
	db := test.SetupPG(t, bootstrap.InitiliazeDB)
	pit := sdk.PipelineInvokerType{
		Author:      "author",
		Description: "description",
		Identifier:  "identifier",
		Name:        "name",
		DockerType: &sdk.PipelineInvokerTypeDocker{
			DockerImage: "img",
			DockerCmd:   "cmd",
		},
	}
	test.NoError(t, InsertInvokerType(db, &pit))

	proj := assets.InsertTestProject(t, db, assets.RandomString(t, 4), assets.RandomString(t, 10))
	app := &sdk.Application{
		Name:       "app",
		ProjectKey: proj.Key,
	}
	pip := &sdk.Pipeline{
		Name:      "pip",
		Type:      sdk.BuildPipeline,
		ProjectID: proj.ID,
	}

	test.NoError(t, application.Insert(db, proj, app))
	test.NoError(t, pipeline.InsertPipeline(db, pip))
	_, err := application.AttachPipeline(db, app.ID, pip.ID)
	test.NoError(t, err)

	i := sdk.PipelineInvoker{
		Type:          pit,
		ApplicationID: app.ID,
		PipelineID:    pip.ID,
		EnvironmentID: sdk.DefaultEnv.ID,
	}

	test.NoError(t, Insert(db, &i))

	ii, err := Load(db, i.UUID)
	test.NoError(t, err)

	t.Log(dump.Sdump(ii))

	test.NoError(t, Delete(db, i.UUID))
	test.NoError(t, DeleteInvokerType(db, pit.ID))

}

func TestDelete(t *testing.T) {
	//coverred by TestInsert
}

func TestLoad(t *testing.T) {
	//coverred by TestInsert
}
