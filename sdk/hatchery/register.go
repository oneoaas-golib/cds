package hatchery

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/facebookgo/httpcontrol"

	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
)

// Create creates hatchery
func Create(h Interface, api, token string, provision int, requestSecondsTimeout int, maxFailures int, insecureSkipVerifyTLS bool, warningSeconds, criticalSeconds, graceSeconds int) {
	Client = &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout:  time.Duration(requestSecondsTimeout) * time.Second,
			MaxTries:        5,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerifyTLS},
		},
	}

	sdk.SetHTTPClient(Client)
	// No user / password, only token used for auth hatchery
	sdk.Options(api, "", "", token)

	if err := h.Init(); err != nil {
		log.Critical("Create> Init error: %s", err)
		os.Exit(10)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Critical("Create> Cannot retrieve hostname: %s", err)
		os.Exit(10)
	}

	go hearbeat(h, token, maxFailures)

	var spawnIds []int64
	var errR error
	for {
		time.Sleep(2 * time.Second)
		if h.Hatchery() == nil || h.Hatchery().ID == 0 {
			log.Debug("Create> continue")
			continue
		}

		spawnIds, errR = routine(h, provision, hostname, time.Now().Unix(), spawnIds, warningSeconds, criticalSeconds, graceSeconds)
		if errR != nil {
			log.Warning("Error on routine: %s", errR)
		}
	}
}

// Register calls CDS API to register current hatchery
func Register(h *sdk.Hatchery, token string) error {
	log.Notice("Register> Hatchery %s", h.Name)

	h.UID = token
	data, errm := json.Marshal(h)
	if errm != nil {
		return errm
	}

	data, code, errr := sdk.Request("POST", "/hatchery", data)
	if errr != nil {
		return errr
	}

	if code >= 300 {
		return fmt.Errorf("Register> HTTP %d", code)
	}

	if err := json.Unmarshal(data, h); err != nil {
		return err
	}

	// Here, h.UID contains token generated by API
	sdk.Authorization(h.UID)

	log.Notice("Register> Hatchery registered with id:%d", h.ID)

	return nil
}

// GenerateName generate a hatchery's name
func GenerateName(add, name string) string {
	if name == "" {
		// Register without declaring model
		var errHostname error
		name, errHostname = os.Hostname()
		if errHostname != nil {
			log.Warning("Cannot retrieve hostname: %s", errHostname)
			name = "cds-hatchery"
		}
		name += "-" + namesgenerator.GetRandomName(0)
	}

	if add != "" {
		name += "-" + add
	}

	return name
}

func hearbeat(m Interface, token string, maxFailures int) {
	var failures int
	for {
		time.Sleep(5 * time.Second)
		if m.Hatchery().ID == 0 {
			log.Notice("hearbeat> Disconnected from CDS engine, trying to register...")
			if err := Register(m.Hatchery(), token); err != nil {
				log.Notice("hearbeat> Cannot register: %s", err)
				checkFailures(maxFailures, failures)
				continue
			}
			if m.Hatchery().ID == 0 {
				log.Critical("hearbeat> Cannot register hatchery. ID %d", m.Hatchery().ID)
				checkFailures(maxFailures, failures)
				continue
			}
			log.Notice("hearbeat> Registered back: ID %d with model ID %d", m.Hatchery().ID, m.Hatchery().Model.ID)
		}

		if _, _, err := sdk.Request("PUT", fmt.Sprintf("/hatchery/%d", m.Hatchery().ID), nil); err != nil {
			log.Notice("heartbeat> cannot refresh beat: %s", err)
			m.Hatchery().ID = 0
			checkFailures(maxFailures, failures)
			continue
		}
		failures = 0
	}
}

func checkFailures(maxFailures, nb int) {
	if nb > maxFailures {
		log.Critical("Too many failures on try register. This hatchery is killed")
		os.Exit(10)
	}
}
