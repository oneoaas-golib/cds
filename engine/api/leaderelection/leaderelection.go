package leaderelection

import (
	"time"

	"strings"

	"github.com/ovh/cds/engine/api/database"
	"github.com/ovh/cds/engine/log"
)

var (
	hostname, cdsname string
	host              *CDSHost
)

// Init initalize the embedded election leader system
func Init(h, s string) {
	hostname = h
	cdsname = s

	go registerRoutine()
	go awolRoutine()
	go electionRoutine()
}

func IsLeader() (bool, error) {
	return false, nil
}

func registerRoutine() {
	for {
		time.Sleep(1 * time.Second)
		_db := database.DB()
		if _db == nil {
			continue
		}
		db := database.DBMap(_db)
		res, err := db.Exec("update leader_election_host set heartbeat = $2 where id = $1", host.ID, time.Now())
		if err != nil {
			log.Critical("leaderelection.registerRoutine> Unable to heartbeat %s : %s", host.Name, err)
			continue
		}
		n, err := res.RowsAffected()
		if err != nil {
			log.Critical("leaderelection.registerRoutine> Unable to heartbeat %s : %s", host.Name, err)
			continue
		}
		//Insert if no rows has been updated
		if n == 0 {
			host = &CDSHost{
				Hostname:  hostname,
				Name:      cdsname,
				IsLeader:  false,
				Heartbeat: time.Now(),
			}
			if err := db.Insert(host); err != nil {
				log.Critical("leaderelection.registerRoutine> Unable to insert host %s : %s", host.Name, err)
			}
		}
	}
}

func electionRoutine() {
	for {
		time.Sleep(1 * time.Second)
		_db := database.DB()
		if _db == nil {
			continue
		}
		db := database.DBMap(_db)
		tx, err := db.Begin()
		if err != nil {
			log.Warning("leaderelection.electionRoutine> Unable to start transaction : %s", err)
			continue
		}
		//Load all hosts
		hosts, err := LoadAll(tx)
		if err != nil {
			log.Warning("leaderelection.electionRoutine> Unable to load hosts : %s", err)
			continue
		}
		//Choose a leader, by name
		var leaderName string
		for _, h := range hosts {
			if strings.Compare(h.Name, leaderName) > 0 {
				leaderName = h.Name
			}
		}
		//Update the leader
		if _, err := db.Exec("update leader_election_host set leader = true where name = $1", host.Name); err != nil {
			log.Critical("leaderelection.registerRoutine> Unable to heartbeat %s : %s", host.Name, err)
			continue
		}
		//Update the other
		if _, err := db.Exec("update leader_election_host set leader = false where name != $1", host.Name); err != nil {
			log.Critical("leaderelection.registerRoutine> Unable to heartbeat %s : %s", host.Name, err)
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Warning("leaderelection.electionRoutine> Unable to commit transaction : %s", err)
			continue
		}
	}
}

func awolRoutine() {
	for {
		time.Sleep(2 * time.Second)
		_db := database.DB()
		if _db == nil {
			continue
		}
		db := database.DBMap(_db)
		//Load all hosts
		hosts, err := LoadAll(db)
		if err != nil {
			log.Warning("leaderelection.awolRoutine> Unable to load hosts : %s", err)
			continue
		}
		for _, h := range hosts {
			//If last Heartbeat is older than 10 seconds ago
			if h.Heartbeat.Before(time.Now().Add(-10 * time.Second)) {
				if _, err := db.Delete(&h); err != nil {
					log.Warning("leaderelection.awolRoutine> Unable to delete hosts : %s %s", h.Name, err)
				}
			}
		}
	}

}
