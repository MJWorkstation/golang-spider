package distribute

import (
	"encoding/json"
	"go-spider/logs"
	"teleport"
)


func MasterApi(n Distributer) teleport.API {
	return teleport.API{
		
		"task": &masterTaskHandle{n},

		
		"log": &masterLogHandle{},
	}
}


type masterTaskHandle struct {
	Distributer
}

func (self *masterTaskHandle) Process(receive *teleport.NetData) *teleport.NetData {
	b, _ := json.Marshal(self.Send(self.CountNodes()))
	return teleport.ReturnData(string(b))
}


type masterLogHandle struct{}

func (*masterLogHandle) Process(receive *teleport.NetData) *teleport.NetData {
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *     [ %s ]    %s", receive.From, receive.Body)
	logs.Log.Informational(" * ")
	return nil
}
