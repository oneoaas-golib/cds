package invoker

import (
	"github.com/ovh/cds/engine/api/database/gorpmapping"
	"github.com/ovh/cds/sdk"
)

// PipelineInvoker is a gorp wrapper around sdk.PipelineInvoker
type PipelineInvoker sdk.PipelineInvoker

// PipelineInvokerType is a gorp wrapper around sdk.PipelineInvokerType
type PipelineInvokerType sdk.PipelineInvokerType

func init() {
	gorpmapping.Register(gorpmapping.New(PipelineInvoker{}, "pipeline_invoker", false, "uuid"))
	gorpmapping.Register(gorpmapping.New(PipelineInvokerType{}, "pipeline_invoker_type", true, "id"))
}
