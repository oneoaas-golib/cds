package source

import (
	"github.com/ovh/cds/engine/api/database/gorpmapping"
	"github.com/ovh/cds/sdk"
)

//EventListener is a gorp wrapper around sdk.EventListener
type EventListener sdk.EventListener

func init() {
	gorpmapping.Register(gorpmapping.New(EventListener{}, "event_listener", false, "uuid"))
}
