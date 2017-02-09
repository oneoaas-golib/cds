package repositoriesmanager

import (
	"regexp"

	"github.com/go-gorp/gorp"

	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
)

// SkipByCommitMessage checks the commit message if application is linked to a repository
func SkipByCommitMessage(db gorp.SqlExecutor, proj *sdk.Project, a *sdk.Application, hash string) bool {
	var match bool
	// Get commit message to check if we have to skip the build
	if a.RepositoriesManager != nil {
		if b, _ := CheckApplicationIsAttached(db, a.RepositoriesManager.Name, proj.Key, a.Name); b && a.RepositoryFullname != "" {
			//Get the RepositoriesManager Client
			client, _ := AuthorizedClient(db, proj.Key, a.RepositoriesManager.Name)
			if client != nil {
				//Get the commit
				commit, err := client.Commit(a.RepositoryFullname, hash)
				if err != nil {
					log.Warning("hook> can't get commit %s from %s on %s : %s", hash, a.RepositoryFullname, a.RepositoriesManager.Name, err)
				}
				//Verify commit message
				match, err = regexp.Match(".*\\[ci skip\\].*|.*\\[cd skip\\].*", []byte(commit.Message))
				if err != nil {
					log.Warning("hook> Cannot check %s/%s for commit %s : %s (%s)", proj.Key, a.Name, hash, commit.Message, err)
				}
				if match {
					log.Notice("hook> Skipping build of %s/%s for commit %s", proj.Key, a.Name, hash)
				}
			}
		} else {
			log.Debug("Application is not attached (%s %s %s)", a.RepositoriesManager.Name, proj.Key, a.Name)
		}
	}
	return match
}
