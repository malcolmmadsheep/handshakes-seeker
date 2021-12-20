package dbservices

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/hash"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type TaskService struct {
	conn        *pgx.Conn
	skipTaskMap sync.Map
}

func NewTaskService(conn *pgx.Conn) *TaskService {
	return &TaskService{
		conn:        conn,
		skipTaskMap: sync.Map{},
	}
}

func scanTask(row pgx.Row) (*services.Task, error) {
	task := services.Task{}

	err := row.Scan(
		&task.Id,
		&task.OriginTaskId,
		&task.DataSource,
		&task.SourceUrl,
		&task.DestUrl,
		&task.Cursor,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (ts *TaskService) ShouldSkipTask(task *services.Task) bool {
	_, shouldSkip := ts.skipTaskMap.Load(task.OriginTaskId)

	return shouldSkip
}

func (ts *TaskService) GenerateId(sourceUrl, destUrl string) string {
	return hash.GetMD5Hash(sourceUrl + destUrl)
}

const getTaskByIdSQL = `
select id, origin_task_id, data_source, source_url, dest_url, cursor 
from tasks_queue
where id = $1;
`

func (ts *TaskService) GetTaskById(id string) (*services.Task, error) {
	task := services.Task{
		Id: id,
	}

	err := ts.conn.QueryRow(
		context.Background(),
		getTaskByIdSQL,
		id,
	).Scan(
		&task.Id,
		&task.OriginTaskId,
		&task.DataSource,
		&task.SourceUrl,
		&task.DestUrl,
		&task.Cursor,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

const createTaskSQL = `
INSERT INTO tasks_queue (id, origin_task_id, data_source, source_url, dest_url, cursor)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`

func (ts *TaskService) CreateNewTask(id, originTaskId, sourceUrl, destUrl, cursor string) (*services.Task, error) {
	task, err := ts.GetTaskById(id)
	if err == nil {
		return task, nil
	}

	_, err = ts.conn.Exec(
		context.Background(),
		createTaskSQL,
		id,
		originTaskId,
		"",
		sourceUrl,
		destUrl,
		cursor,
	)
	if err != nil {
		return nil, err
	}

	return &services.Task{
		Id:           id,
		OriginTaskId: originTaskId,
		SourceUrl:    sourceUrl,
		DestUrl:      destUrl,
		Cursor:       cursor,
	}, nil
}

const getNEarliestTasksSQL = `
select id, origin_task_id, data_source, source_url, dest_url, cursor 
from tasks_queue
order by created_at
limit $1;
`

func (ts *TaskService) GetNEarliestTasks(n uint) ([]*services.Task, error) {
	tasks := make([]*services.Task, 0, n)

	rows, err := ts.conn.Query(context.Background(), getNEarliestTasksSQL, n)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

const deleteTaskByIdSQL = `
delete from tasks_queue
where id = $1;
`

func (ts *TaskService) DeleteTaskById(id string) error {
	_, err := ts.conn.Exec(context.Background(), deleteTaskByIdSQL, id)

	return err
}

const deleteAllTasksByOriginIdSQL = `
delete from tasks_queue
where origin_task_id = $1;
`

func (ts *TaskService) DeleteAllTasksWithOrigin(originId string) error {
	_, err := ts.conn.Exec(context.Background(), deleteAllTasksByOriginIdSQL, originId)
	if err != nil {
		return err
	}

	ts.skipTaskMap.Store(originId, true)

	return nil
}
