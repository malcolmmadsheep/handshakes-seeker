package dbservices

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/hash"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type TaskService struct {
	conn        *pgxpool.Pool
	skipTaskMap sync.Map
}

func NewTaskService(conn *pgxpool.Pool) *TaskService {
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
	count, contains := ts.skipTaskMap.Load(task.OriginTaskId)

	return contains && count == 0
}

func (ts *TaskService) addTaskCount(id string, n int) int {
	count, _ := ts.skipTaskMap.Load(id)

	if count == nil {
		ts.skipTaskMap.Store(id, n)
		return n
	}

	newCount := count.(int) + n

	return newCount
}

func (ts *TaskService) incrementTaskCount(id string) int {
	return ts.addTaskCount(id, 1)
}

func (ts *TaskService) decrementTaskCount(id string) int {
	return ts.addTaskCount(id, -1)
}

func (ts *TaskService) GenerateId(sourceUrl, destUrl string) string {
	return hash.GetMD5Hash(sourceUrl + destUrl)
}

const getTaskByIdSQL = `
select id, origin_task_id, data_source, source_url, dest_url, cursor, requests_count
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
		&task.RequestsCount,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

const updateTaskRequestCountSQL = `
UPDATE tasks_queue
set requests_count = requests_count + $1
where origin_task_id = $2
RETURNING requests_count;	
`

func (ts *TaskService) UpdateTaskRequestsCount(originTaskId string, n int) (int, error) {
	count := 0

	err := ts.conn.QueryRow(context.Background(), updateTaskRequestCountSQL, n, originTaskId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

const createTaskSQL = `
INSERT INTO tasks_queue (id, origin_task_id, data_source, source_url, dest_url, cursor, requests_count)
VALUES ($1, $2, $3, $4, $5, $6, $7);
`

func (ts *TaskService) CreateNewTask(id, originTaskId, sourceUrl, destUrl, cursor string, requestsCount int) (*services.Task, error) {
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
		requestsCount,
	)
	if err != nil {
		return nil, err
	}

	ts.incrementTaskCount(id)

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
where id = $1 and origin_task_id = $2;
`

func (ts *TaskService) DeleteTaskByIds(id string, originId string) error {
	_, err := ts.conn.Exec(context.Background(), deleteTaskByIdSQL, id, originId)

	return err
}

const deleteAllTasksByOriginIdSQL = `
delete from tasks_queue
where origin_task_id = $1;
`

func (ts *TaskService) DeleteAllTasksWithOrigin(originId string) error {
	ts.decrementTaskCount(originId)

	_, err := ts.conn.Exec(context.Background(), deleteAllTasksByOriginIdSQL, originId)
	if err != nil {
		return err
	}

	return nil
}
