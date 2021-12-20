package services

type TaskBody struct {
	SourceUrl string `json:"source_url"`
	DestUrl   string `json:"dest_url"`
	Cursor    string `json:"cursor"`
}

type Task struct {
	Id           string `json:"id"`
	OriginTaskId string `json:"origin_task_id"`
	DataSource   string `json:"data_source"`
	SourceUrl    string `json:"source_url"`
	DestUrl      string `json:"dest_url"`
	Cursor       string `json:"cursor"`
}

type TaskService interface {
	ShouldSkipTask(*Task) bool

	GenerateId(sourceUrl, destUrl string) string
	GetTaskById(id string) (*Task, error)
	CreateNewTask(id, originalTaskId, sourceUrl, destUrl, cursor string) (*Task, error)
	GetNEarliestTasks(uint) ([]*Task, error)
	DeleteTaskById(string) error
	DeleteAllTasksWithOrigin(string) error
}
