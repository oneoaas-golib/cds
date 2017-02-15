package sdk

import "time"

// PipelineInvokeEvent is an event
type PipelineInvokeEvent struct {
	Date         time.Time              `json:"date"`
	Source       string                 `json:"source"`
	Payload      map[string]interface{} `json:"payload"`
	ListenerUUID string                 `json:"listener" db:"listener_uuid"`
}

// PipelineInvoker is an event listener
type PipelineInvoker struct {
	UUID          string              `json:"uuid" db:"uuid"`
	ApplicationID int64               `json:"-" db:"application_id"`
	PipelineID    int64               `json:"-" db:"pipeline_id"`
	EnvironmentID int64               `json:"-" db:"environment_id"`
	Application   Application         `json:"application" db:"-"`
	Pipeline      Pipeline            `json:"pipeline" db:"-"`
	Environment   Environment         `json:"environment" db:"-"`
	Type          PipelineInvokerType `json:"type" db:"-"`
	TypeID        int64               `json:"-" db:"pipeline_invoker_type_id"`
}

// PipelineInvokerType is a type of listener
type PipelineInvokerType struct {
	ID          int64                      `json:"id" db:"id"`
	Name        string                     `json:"name" db:"name"`
	Author      string                     `json:"author" db:"author"`
	Description string                     `json:"description" db:"description"`
	Identifier  string                     `json:"identifier" db:"identifier"`
	DockerType  *PipelineInvokerTypeDocker `json:"type,omitempty" db:"-"`
	LocalType   *PipelineInvokerTypeLocal  `json:"type,omitempty" db:"-"`
}

// PipelineInvokerTypeDocker for listener of type docker
type PipelineInvokerTypeDocker struct {
	DockerImage string `json:"docker_image" db:"docker_image"`
	DockerCmd   string `json:"docker_cmd" db:"docker_cmd"`
}

// PipelineInvokerTypeLocal for listener of type local
type PipelineInvokerTypeLocal struct {
	LocalCmd   string `json:"local_cmd"`
	Size       int64  `json:"size"`
	Perm       uint32 `json:"perm"`
	MD5Sum     string `json:"md5sum"`
	ObjectPath string `json:"object_path"`
}
