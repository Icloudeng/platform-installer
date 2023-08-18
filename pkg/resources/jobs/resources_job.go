package jobs

import (
	"context"
	"smatflow/platform-installer/pkg/database"
	"smatflow/platform-installer/pkg/queue"
	"smatflow/platform-installer/pkg/resources/db"
	"smatflow/platform-installer/pkg/resources/websocket"

	goqueue "github.com/golang-queue/queue"
)

type ResourcesJob struct {
	Ref           string
	Task          goqueue.TaskFunc
	PostBody      interface{}
	ResourceState bool
	Method        string
	Description   string
	Group         string
	Handler       string
}

func ResourcesJobTask(task ResourcesJob) *database.Job {
	// Create new JOB
	job := db.JobCreate(db.JobCreateParam{
		Ref:         task.Ref,
		PostBody:    task.PostBody,
		Description: task.Description,
		Group:       task.Group,
		Handler:     task.Handler,
		Method:      task.Method,
		Status:      database.JOB_STATUS_IDLE,
	})

	//Emit ws events
	websocket.EmitJobEvent(job)

	queue.Queue.QueueTask(func(ctx context.Context) error {
		res_state := &database.ResourcesState{}

		job = db.JobUpdateStatus(job, database.JOB_STATUS_RUNNING)
		//Emit ws events
		websocket.EmitJobEvent(job)

		// Listen to Redis Provisining events
		close := redis_pub_listeners(task.Ref)
		defer close()

		// Create Resource State
		if task.ResourceState {
			res_state = db.ResourceStateCreate(task.Ref, *job)
		}

		// Run task
		err := task.Task(ctx)

		if err == nil && task.ResourceState {
			db.ResourceStatePutTerraformState(res_state)
		}

		if err == nil {
			job = db.JobUpdateStatus(job, database.JOB_STATUS_COMPLETED)
		} else {
			job = db.JobUpdateLogs(job, err.Error())
			job = db.JobUpdateStatus(job, database.JOB_STATUS_FAILED)
		}

		//Emit ws events
		websocket.EmitJobEvent(job)

		// Allocate memory for resource db backup
		go db.CreateNewResourcesBackup()

		return nil
	})

	return job
}
