package distribute


type TaskJar struct {
	Tasks chan *Task
}

func NewTaskJar() *TaskJar {
	return &TaskJar{
		Tasks: make(chan *Task, 1024),
	}
}


func (self *TaskJar) Push(task *Task) {
	id := len(self.Tasks)
	task.Id = id
	self.Tasks <- task
}


func (self *TaskJar) Pull() *Task {
	return <-self.Tasks
}


func (self *TaskJar) Len() int {
	return len(self.Tasks)
}


func (self *TaskJar) Send(clientNum int) Task {
	return *<-self.Tasks
}


func (self *TaskJar) Receive(task *Task) {
	self.Tasks <- task
}
