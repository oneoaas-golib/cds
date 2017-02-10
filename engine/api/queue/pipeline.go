package queue

import (
	"time"

	"github.com/go-gorp/gorp"

	"github.com/ovh/cds/engine/api/cache"
	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/repositoriesmanager"
	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
)

// PipelineRunEventListener si the main goroutine
func PipelineRunEventListener(DBFunc func() gorp.SqlExecutor) {
	// If this goroutine exits, then it's a crash
	defer log.Fatalf("Goroutine of NewPipelineBuildEventListener exited - Exit CDS Engine")

	for {
		time.Sleep(10 * time.Millisecond)

		//Check if CDS is in maintenance mode
		var m bool
		cache.Get("maintenance", &m)
		if m {
			log.Warning("⚠ CDS maintenance in ON")
			time.Sleep(30 * time.Second)
		}

		var e sdk.PipelineRunEvent
		cache.Dequeue("events_pipelinerun", &e)

		db := DBFunc()
		if db != nil && !m {
			if err := newPipelineRunFromEvent(db, e); err != nil {
				log.Critical("NewPipelineBuildEventListener> err while processing %s : %v", err, e)
			}
			continue
		} else {
			cache.Enqueue("events_pipelinerun", e)
		}
	}
}

func newPipelineRunFromEvent(db gorp.SqlExecutor, e sdk.PipelineRunEvent) error {
	return nil
}

// NewPipelineBuild add a new pipeline in queue
// Pipeline should be full loaded
func NewPipelineBuild(db gorp.SqlExecutor, proj *sdk.Project, pip *sdk.Pipeline, app *sdk.Application, env *sdk.Environment, t sdk.PipelineBuildTrigger, args ...sdk.Parameter) error {
	//Check if we have a git.hash
	p := sdk.ParameterFind(args, "git.hash")
	if p != nil {
		if repositoriesmanager.SkipByCommitMessage(db, proj, app, p.Value) {
			return nil
		}
	}

	//Insert the pipeline build
	if _, err := pipeline.InsertPipelineBuild(db, proj, pip, app, nil, args, env, 0, t); err != nil {
		return err
	}

	return nil
}
