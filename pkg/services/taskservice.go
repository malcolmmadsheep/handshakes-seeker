package services

type Task struct {
	Id   uint
	Body []byte
}

type TaskService interface {
	GetTaskByBody([]byte) (*Task, error)
	CreateNewTask([]byte) (*Task, error)
	GetNEarliestTasks(uint) ([]*Task, error)
	DeleteTaskById(uint) error
}
