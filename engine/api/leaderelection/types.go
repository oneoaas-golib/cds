package leaderelection

import "time"

// CDSHost is the entry for leader election system
type CDSHost struct {
	ID        int       `db:"id"`
	Hostname  string    `db:"hostname"`
	Name      string    `db:"name"`
	IsLeader  bool      `db:"is_leader"`
	Heartbeat time.Time `db:"heartbeat"`
}
