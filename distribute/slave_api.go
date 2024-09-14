package distribute

import (
	"encoding/json"
	"go-spider/logs"
	"teleport"
)


func SlaveApi(n Distributer) teleport.API {
	return teleport.API{
		
		"task": &slaveTaskHandle{n},
	}
}


type slaveTaskHandle struct {
	Distributer
}

func (self *slaveTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	t := &Task{}
	err := json.Unmarshal([]byte(receive.Body.(string)), t)
	if err != nil {
		logs.Log.Error("json解码失败 %v", receive.Body)
		return nil
	}
	self.Receive(t)
	return nil
}
