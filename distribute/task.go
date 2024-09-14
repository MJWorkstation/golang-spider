package distribute


type Task struct {
	Id             int
	Spiders        []map[string]string 
	ThreadNum      int                 
	Pausetime      int64               
	OutType        string              
	DockerCap      int                 
	DockerQueueCap int                 
	SuccessInherit bool                
	FailureInherit bool                
	Limit          int64               
	ProxyMinute    int64               
	
	Keyins string 
}
