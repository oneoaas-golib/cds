package invoker

import (
	"testing"

	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/sdk"
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
	test.NoError(t, DeleteInvokerType(_db, pit.ID))
}

func TestUpdateInvokerType(t *testing.T) {
	_ = test.SetupPG(t)
}

func TestDeleteInvokerType(t *testing.T) {
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
