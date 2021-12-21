package services

type PathStatus uint

const (
	PathStatusNotStarted PathStatus = iota
	PathStatusInProgress
	PathStatusFound
	PathStatusNotFound
	PathStatusCancelled
)

func (s PathStatus) String() string {
	switch s {
	case PathStatusNotStarted:
		return "not_started"
	case PathStatusInProgress:
		return "in_progress"
	case PathStatusFound:
		return "found"
	case PathStatusNotFound:
		return "not_found"
	case PathStatusCancelled:
		return "cancelled"
	}
	return "unknown"
}

type Path struct {
	Id        uint   `json:"id"`
	SourceUrl string `json:"source_url"`
	DestUrl   string `json:"dest_url"`
	TaskHash  string
	Status    string `json:"status"`
	Trace     string `json:"trace"`
}

type PathShapeForBulk struct {
	TaskId    string
	SourceUrl string
	DestUrl   string
	Trace     string
}

type PathService interface {
	GetPathByTaskId(taskId string) (*Path, error)
	CreateNewPath(task *Task) (*Path, error)
	CreateFoundPath(taskId, sourceUrl, destUrl, trace string) (*Path, error) // make it batch
	BulkCreateFoundPaths([]PathShapeForBulk) error                           // make it batch
	UpdatePathStatusByTaskId(taskId string, status PathStatus) error
	BuildFullTraceAndUpdate(path *Path) (*Path, error)
}
