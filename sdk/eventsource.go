package sdk

import "time"

// PipelineRunEvent is an event
type PipelineRunEvent struct {
	Date         time.Time              `json:"date"`
	Source       string                 `json:"source"`
	Payload      map[string]interface{} `json:"payload"`
	ListenerUUID string                 `json:"listener" db:"listener_uuid"`
}

// EventListener is an event listener
type EventListener struct {
	UUID          string      `json:"uuid" db:"uuid"`
	ApplicationID int64       `json:"-" db:"application_id"`
	PipelineID    int64       `json:"-" db:"pipeline_id"`
	EnvironmentID int64       `json:"-" db:"environment_id"`
	Application   Application `json:"application" db:"-"`
	Pipeline      Pipeline    `json:"pipeline" db:"-"`
	Environment   Environment `json:"environment" db:"-"`
}
