package dbservices

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type PathService struct {
	conn *pgxpool.Pool
}

func NewPathService(conn *pgxpool.Pool) *PathService {
	return &PathService{
		conn,
	}
}

const getPathByTaskIdSQL = `
select id, source_url, destination_url, status, trace
from paths
where task_hash = $1;
`

func (ps *PathService) GetPathByTaskId(taskId string) (*services.Path, error) {
	path := services.Path{}

	err := ps.conn.QueryRow(
		context.Background(),
		getPathByTaskIdSQL,
		taskId,
	).Scan(
		&path.Id,
		&path.SourceUrl,
		&path.DestUrl,
		&path.Status,
		&path.Trace,
	)
	if err != nil {
		return nil, err
	}

	return &path, nil
}

const createNewPathSQL = `
insert into paths (data_source, task_hash, source_url, destination_url, status, trace)
values ($1, $2, $3, $4, $5, $6)
returning id;
`

func (ps *PathService) createNewPath(taskId, sourceUrl, destUrl, trace string, status services.PathStatus) (*services.Path, error) {
	path, err := ps.GetPathByTaskId(taskId)
	if err == nil {
		return path, nil
	} else if err != pgx.ErrNoRows {
		return nil, err
	}

	newPath := services.Path{
		SourceUrl: sourceUrl,
		DestUrl:   destUrl,
		Status:    status.String(),
		TaskHash:  taskId,
		Trace:     trace,
	}

	var id uint = 0
	err = ps.conn.QueryRow(
		context.Background(),
		createNewPathSQL,
		"",
		newPath.TaskHash,
		newPath.SourceUrl,
		newPath.DestUrl,
		newPath.Status,
		newPath.Trace,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	newPath.Id = id

	return &newPath, nil
}

func (ps *PathService) CreateNewPath(task *services.Task) (*services.Path, error) {
	return ps.createNewPath(task.Id, task.SourceUrl, task.DestUrl, "", services.PathStatusInProgress)
}

const updatePathStatusSQL = `
update paths
set status = $1
where task_hash = $2;
`

func (ps *PathService) UpdatePathStatusByTaskId(taskId string, status services.PathStatus) error {
	_, err := ps.conn.Exec(context.Background(), updatePathStatusSQL, status.String(), taskId)

	return err
}

func (ps *PathService) BuildFullTraceAndUpdate(pathId uint) (string, error) {
	return "<path_should_be_here>", nil
}

func (ps *PathService) CreateFoundPath(taskId, sourceUrl, destUrl, trace string) (*services.Path, error) {
	return ps.createNewPath(taskId, sourceUrl, destUrl, trace, services.PathStatusFound)
}

func (ps *PathService) BulkCreateFoundPaths(shapes []services.PathShapeForBulk) error {
	sb := strings.Builder{}

	sb.WriteString("insert into paths (task_hash, source_url, destination_url, status, trace) values ($1, $2, $3, $4, $5)")

	for _, shape := range shapes {
		sb.WriteString(fmt.Sprintf(`(%s, %s, %s, %s, %s)`, shape.TaskId, shape.SourceUrl, shape.DestUrl, services.PathStatusFound, shape.Trace))
	}

	_, err := ps.conn.Exec(context.Background(), sb.String())

	return err
}
