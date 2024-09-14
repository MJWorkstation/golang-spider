package distribute


type Distributer interface {
	
	Send(clientNum int) Task
	
	Receive(task *Task)
	
	CountNodes() int
}
