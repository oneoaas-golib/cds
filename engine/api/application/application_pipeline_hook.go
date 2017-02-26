package application

import (
	"github.com/go-gorp/gorp"

	"github.com/ovh/cds/engine/api/pipeline"
	"github.com/ovh/cds/engine/api/repositoriesmanager"
	"github.com/ovh/cds/sdk"
)

// TriggerPipeline linked to received hook
func TriggerPipeline(tx gorp.SqlExecutor, h sdk.Hook, branch string, hash string, author string, p *sdk.Pipeline, projectData *sdk.Project) (bool, error) {

	// Create pipeline args
	var args []sdk.Parameter
	args = append(args, sdk.Parameter{
		Name:  "git.branch",
		Value: branch,
	})
	args = append(args, sdk.Parameter{
		Name:  "git.hash",
		Value: hash,
	})
	args = append(args, sdk.Parameter{
		Name:  "git.author",
		Value: author,
	})
	args = append(args, sdk.Parameter{
		Name:  "git.repository",
		Value: h.Repository,
	})
	args = append(args, sdk.Parameter{
		Name:  "git.project",
		Value: h.Project,
	})

	// Load pipeline Argument
	parameters, err := pipeline.GetAllParametersInPipeline(tx, p.ID)
	if err != nil {
		return false, err
	}
	p.Parameter = parameters

	// get application
	a, err := LoadByID(tx, h.ApplicationID, nil, LoadOptions.WithRepositoryManager, LoadOptions.WithVariablesWithClearPassword)
	if err != nil {
		return false, err
	}
	applicationPipelineArgs, err := GetAllPipelineParam(tx, h.ApplicationID, p.ID)
	if err != nil {
		return false, err
	}

	trigger := sdk.PipelineBuildTrigger{
		ManualTrigger:    false,
		VCSChangesBranch: branch,
		VCSChangesHash:   hash,
		VCSChangesAuthor: author,
	}

	if repositoriesmanager.SkipByCommitMessage(tx, projectData, a, hash) {
		return false, nil
	}

	// FIXME add possibility to trigger a pipeline on a specific env
	_, err = pipeline.InsertPipelineBuild(tx, projectData, p, a, applicationPipelineArgs, args, &sdk.DefaultEnv, 0, trigger)
	if err != nil {
		return false, err
	}

	return true, nil
}
