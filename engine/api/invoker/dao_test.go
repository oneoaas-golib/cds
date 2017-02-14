package invoker

import (
	"testing"

	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/sdk"
	"github.com/stretchr/testify/assert"
)

func TestInsertInvokerType(t *testing.T) {
	_db := test.SetupPG(t)

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
	test.NoError(t, InsertInvokerType(_db, &pit))

	pit1, err := LoadInvokerType(_db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(_db, pit.ID))
}

func TestUpdateInvokerType(t *testing.T) {
	_db := test.SetupPG(t)

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
	test.NoError(t, InsertInvokerType(_db, &pit))

	pit1, err := LoadInvokerType(_db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)

	pit.Description = "description2"
	pit.LocalType = &sdk.PipelineInvokerTypeLocal{
		LocalCmd: "local",
	}
	pit.DockerType = nil

	test.NoError(t, UpdateInvokerType(_db, &pit))
	pit1, err = LoadInvokerType(_db, pit.ID)

	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(_db, pit.ID))
}

func TestDeleteInvokerType(t *testing.T) {
	_db := test.SetupPG(t)
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
	test.NoError(t, InsertInvokerType(_db, &pit))

	pit1, err := LoadInvokerType(_db, pit.ID)
	test.NoError(t, err)
	test.EqualValues(t, pit, pit1)
	test.NoError(t, DeleteInvokerType(_db, pit.ID))

	_, err = LoadInvokerType(_db, pit.ID)
	assert.Equal(t, sdk.ErrPipelineInvokerNotFound, err)

	_ = test.SetupPG(t)
}

func TestLoadInvokerType(t *testing.T) {
	_ = test.SetupPG(t)
}

func TestLoadAllInvokerTypes(t *testing.T) {
	_ = test.SetupPG(t)
}

func TestInsert(t *testing.T) {
	_ = test.SetupPG(t)
}

func TestDelete(t *testing.T) {
	_ = test.SetupPG(t)
}

func TestLoad(t *testing.T) {
	_ = test.SetupPG(t)
}
