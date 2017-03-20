package marathon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/gambol99/go-marathon"
	"github.com/spf13/viper"

	"github.com/ovh/cds/engine/log"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/hatchery"
)

var hatcheryMarathon *HatcheryMarathon

type marathonPOSTAppParams struct {
	DockerImage    string
	ForcePullImage bool
	APIEndpoint    string
	WorkerKey      string
	WorkerName     string
	WorkerModelID  int64
	HatcheryID     int64
	JobID          int64
	MarathonID     string
	MarathonVHOST  string
	MarathonLabels string
	Memory         int
	WorkerTTL      int
}

const marathonPOSTAppTemplate = `
{
    "container": {
        "docker": {
            "forcePullImage": {{.ForcePullImage}},
            "image": "{{.DockerImage}}",
            "network": "BRIDGE",
            "portMapping": []
				},
        "type": "DOCKER"
    },
		"cmd": "rm -f worker && curl ${CDS_API}/download/worker/$(uname -m) -o worker &&  chmod +x worker && exec ./worker",
		"cpus": 0.5,
    "env": {
        "CDS_API": "{{.APIEndpoint}}",
        "CDS_KEY": "{{.WorkerKey}}",
        "CDS_NAME": "{{.WorkerName}}",
        "CDS_MODEL": "{{.WorkerModelID}}",
        "CDS_HATCHERY": "{{.HatcheryID}}",
        "CDS_BOOKED_JOB_ID": "{{.JobID}}",
        "CDS_SINGLE_USE": "1",
        "CDS_TTL" : "{{.WorkerTTL}}"
    },
    "id": "{{.MarathonID}}/{{.WorkerName}}",
    "instances": 1,
    "ports": [],
    "mem": {{.Memory}},
    "labels": {{.MarathonLabels}}
}
`

// HatcheryMarathon implements HatcheryMode interface for mesos mode
type HatcheryMarathon struct {
	hatch *sdk.Hatchery
	token string

	client marathon.Marathon

	marathonHost     string
	marathonUser     string
	marathonPassword string

	marathonID           string
	marathonVHOST        string
	marathonLabelsString string
	marathonLabels       map[string]string

	defaultMemory      int
	workerTTL          int
	workerSpawnTimeout int
}

// ID must returns hatchery id
func (m *HatcheryMarathon) ID() int64 {
	if m.hatch == nil {
		return 0
	}
	return m.hatch.ID
}

//Hatchery returns hatchery instance
func (m *HatcheryMarathon) Hatchery() *sdk.Hatchery {
	return m.hatch
}

// KillWorker deletes an application on mesos via marathon
func (m *HatcheryMarathon) KillWorker(worker sdk.Worker) error {
	appID := path.Join(hatcheryMarathon.marathonID, worker.Name)
	log.Notice("KillWorker> Killing %s", appID)

	_, err := m.client.DeleteApplication(appID, true)
	return err
}

// CanSpawn return wether or not hatchery can spawn model
// requirements services are not supported
func (m *HatcheryMarathon) CanSpawn(model *sdk.Model, job *sdk.PipelineBuildJob) bool {
	if model.Type != sdk.Docker {
		return false
	}
	//Service requirement are not supported
	for _, r := range job.Job.Action.Requirements {
		if r.Type == sdk.ServiceRequirement {
			return false
		}
	}

	return true
}

// SpawnWorker creates an application on mesos via marathon
// requirements services are not supported
func (m *HatcheryMarathon) SpawnWorker(model *sdk.Model, job *sdk.PipelineBuildJob) error {
	if model.Type != sdk.Docker {
		return fmt.Errorf("spawnWorker> model %s not handled for hatchery marathon", model.Type)
	}

	if job != nil {
		log.Notice("spawnWorker> spawning worker %s (%s) for job %d", model.Name, model.Image, job.ID)
	} else {
		log.Notice("spawnWorker> spawning worker %s (%s)", model.Name, model.Image)
	}

	deployments, errd := m.client.Deployments()
	if errd != nil {
		return errd
	}
	// Do not DOS marathon, if deployment queue is longer than 10, wait
	if len(deployments) >= 10 {
		log.Notice("spawnWorker> %d item in deployment queue, waiting", len(deployments))
		time.Sleep(2 * time.Second)
		return nil
	}

	apps, err := m.listApplications(m.marathonID)
	if err != nil {
		return err
	}
	if len(apps) >= viper.GetInt("max-worker") {
		return fmt.Errorf("spawnWorker> max number of containers reached, aborting")
	}

	return m.spawnMarathonDockerWorker(model, m.hatch.ID, job)
}

func (m *HatcheryMarathon) listApplications(idPrefix string) ([]string, error) {
	values := url.Values{}
	values.Set("embed", "apps.counts")
	values.Set("id", hatcheryMarathon.marathonID)
	return m.client.ListApplications(values)
}

// WorkerStarted returns the number of instances of given model started but
// not necessarily register on CDS yet
func (m *HatcheryMarathon) WorkerStarted(model *sdk.Model) int {
	apps, err := m.listApplications(hatcheryMarathon.marathonID)
	if err != nil {
		return 0
	}

	var x int
	for _, app := range apps {
		if strings.Contains(app, strings.ToLower(model.Name)) {
			x++
		}
	}

	return x
}

// Init only starts killing routine of worker not registered
func (m *HatcheryMarathon) Init() error {
	// Register without declaring model
	m.hatch = &sdk.Hatchery{
		Name: hatchery.GenerateName("marathon", viper.GetBool("random-name")),
		UID:  viper.GetString("token"),
	}

	if err := hatchery.Register(m.hatch, viper.GetString("token")); err != nil {
		log.Warning("Cannot register hatchery: %s", err)
	}

	// Start cleaning routines
	m.startKillAwolWorkerRoutine()
	return nil
}

func (m *HatcheryMarathon) marathonConfig(model *sdk.Model, hatcheryID int64, job *sdk.PipelineBuildJob, memory int) ([]byte, error) {
	tmpl, err := template.New("marathonPOST").Parse(marathonPOSTAppTemplate)
	if err != nil {
		return nil, err
	}

	m.marathonLabels["hatchery"] = fmt.Sprintf("%d", hatcheryID)

	labels, err := json.Marshal(m.marathonLabels)
	if err != nil {
		log.Critical("marathonConfig> Invalid labels : %s", err)
		return nil, err
	}

	var jobID int64
	if job != nil {
		jobID = job.ID
	}

	params := marathonPOSTAppParams{
		ForcePullImage: strings.HasSuffix(model.Image, ":latest"),
		DockerImage:    model.Image,
		APIEndpoint:    sdk.Host,
		WorkerKey:      m.token,
		WorkerName:     fmt.Sprintf("%s-%s", strings.ToLower(model.Name), strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)),
		WorkerModelID:  model.ID,
		HatcheryID:     hatcheryID,
		JobID:          jobID,
		MarathonID:     m.marathonID,
		MarathonVHOST:  m.marathonVHOST,
		Memory:         memory * 110 / 100,
		MarathonLabels: string(labels),
		WorkerTTL:      m.workerTTL,
	}

	buffer := &bytes.Buffer{}
	if err := tmpl.Execute(buffer, params); err != nil {
		log.Critical("Unable to execute marathon template : %s", err)
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m *HatcheryMarathon) spawnMarathonDockerWorker(model *sdk.Model, hatcheryID int64, job *sdk.PipelineBuildJob) error {
	// Estimate needed memory, we will set 110% of required memory
	memory := m.defaultMemory
	//Check if there is a memory requirement
	//if there is a service requirement: exit
	if job != nil {
		for _, r := range job.Job.Action.Requirements {
			if r.Name == sdk.ServiceRequirement {
				return fmt.Errorf("spawnMarathonDockerWorker> Service requirement not supported")
			}

			if r.Type == sdk.MemoryRequirement {
				var err error
				memory, err = strconv.Atoi(r.Value)
				if err != nil {
					log.Warning("spawnMarathonDockerWorker> Unable to parse memory requirement %s :s", memory, err)
					return err
				}
			}
		}
	}

	buffer, errm := m.marathonConfig(model, hatcheryID, job, memory)
	if errm != nil {
		return errm
	}

	application := &marathon.Application{}
	if err := json.Unmarshal(buffer, application); err != nil {
		log.Warning("spawnMarathonDockerWorker> configuration file parse error err:%s", err)
		return err
	}

	if _, err := m.client.CreateApplication(application); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		t0 := time.Now()
		for t := range ticker.C {
			delta := math.Floor(t.Sub(t0).Seconds())
			log.Debug("spawnMarathonDockerWorker> application %s spawning in progress [%d seconds] please wait...", application.ID, int(delta))
		}
	}()

	log.Debug("Application %s spawning in progress, please wait...", application.ID)
	deployments, err := m.client.ApplicationDeployments(application.ID)
	if err != nil {
		ticker.Stop()
		return fmt.Errorf("spawnMarathonDockerWorker> failed to list deployments: %s", err.Error())
	}

	if len(deployments) == 0 {
		ticker.Stop()
		return nil
	}

	wg := &sync.WaitGroup{}
	var done bool
	var successChan = make(chan bool, len(deployments))
	for _, deploy := range deployments {
		wg.Add(1)
		go func(id string) {
			go func() {
				time.Sleep((time.Duration(m.workerSpawnTimeout) + 1) * time.Second)
				if done {
					return
				}
				// try to delete deployment
				log.Debug("timeout (%d) on deployment %s", m.workerSpawnTimeout, id)
				if _, err := m.client.DeleteDeployment(id, true); err != nil {
					log.Warning("Error on delete timeouted deployment %s: %s", id, err.Error())
				}
				ticker.Stop()
				successChan <- false
				wg.Done()
			}()

			if err := m.client.WaitOnDeployment(id, time.Duration(m.workerSpawnTimeout)*time.Second); err != nil {
				log.Warning("Error on deployment %s: %s", id, err.Error())
				ticker.Stop()
				successChan <- false
				wg.Done()
				return
			}

			log.Debug("Deployment %s succeeded", id)
			ticker.Stop()
			successChan <- true
			wg.Done()
		}(deploy.DeploymentID)
	}

	wg.Wait()

	var success = true
	for b := range successChan {
		success = success && b
		if len(successChan) == 0 {
			break
		}
	}
	ticker.Stop()
	close(successChan)
	done = true

	if success {
		return nil
	}

	return fmt.Errorf("Error while deploying worker")
}

func (m *HatcheryMarathon) startKillAwolWorkerRoutine() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := m.killDisabledWorkers(); err != nil {
				log.Warning("Cannot kill awol workers: %s", err)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			if err := m.killAwolWorkers(); err != nil {
				log.Warning("Cannot kill awol workers: %s", err)
			}
		}
	}()
}

func (m *HatcheryMarathon) killDisabledWorkers() error {
	workers, err := sdk.GetWorkers()
	if err != nil {
		return err
	}

	apps, err := m.listApplications(hatcheryMarathon.marathonID)
	if err != nil {
		return err
	}

	for _, w := range workers {
		if w.Status != sdk.StatusDisabled {
			continue
		}

		// check that there is a worker matching
		for _, app := range apps {
			if strings.HasSuffix(app, w.Name) {
				log.Notice("killing disabled worker %s", app)
				if _, err := m.client.DeleteApplication(app, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (m *HatcheryMarathon) killAwolWorkers() error {
	workers, err := sdk.GetWorkers()
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("embed", "apps.counts")
	values.Set("id", hatcheryMarathon.marathonID)

	apps, err := m.client.Applications(values)
	if err != nil {
		return err
	}

	var found bool
	// then for each RUNNING marathon application
	for _, app := range apps.Apps {
		// Worker is deploying, leave him alone
		if app.TasksRunning == 0 {
			continue
		}
		t, err := time.Parse(time.RFC3339, app.Version)
		if err != nil {
			log.Warning("Cannot parse last update: %s", err)
			break
		}

		// check that there is a worker matching
		found = false
		for _, w := range workers {
			if strings.HasSuffix(app.ID, w.Name) && w.Status != sdk.StatusDisabled {
				found = true
				break
			}
		}

		// then if it's not found, kill it !
		if !found && time.Since(t) > 1*time.Minute {
			log.Notice("killing awol worker %s", app.ID)
			if _, err := m.client.DeleteApplication(app.ID, true); err != nil {
				return err
			}
		}
	}

	return nil
}
