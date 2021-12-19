package dbservices

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type TaskService struct {
	conn *pgx.Conn
}

func NewTaskService(conn *pgx.Conn) *TaskService {
	return &TaskService{
		conn,
	}
}

const getTaskByBodySQL = `
select id from tasks_queue 
where body = $1;
`

func (ts *TaskService) GetTaskByBody(body []byte) (*services.Task, error) {
	task := services.Task{
		Body: body,
	}

	err := ts.conn.QueryRow(context.Background(), getTaskByBodySQL, body).Scan(&task.Id)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

const createTaskSQL = `
INSERT INTO tasks_queue (body)
VALUES ($1)
RETURNING id;
`

func (ts *TaskService) CreateNewTask(body []byte) (*services.Task, error) {
	var taskId uint
	err := ts.conn.QueryRow(context.Background(), createTaskSQL, body).Scan(&taskId)
	if err != nil {
		return nil, err
	}

	return &services.Task{
		Id:   taskId,
		Body: body,
	}, nil
}

const getNEarliestTasksSQL = `
select id, body from tasks_queue
order by id
limit $1;
`

func (ts *TaskService) GetNEarliestTasks(n uint) ([]*services.Task, error) {
	tasks := make([]*services.Task, 0, n)

	rows, err := ts.conn.Query(context.Background(), getNEarliestTasksSQL, n)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		task := &services.Task{}

		err = rows.Scan(&task.Id, &task.Body)
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

func (ts *TaskService) DeleteTaskById(id uint) error {
	_, err := ts.conn.Exec(context.Background(), deleteTaskByIdSQL, id)

	return err
}
