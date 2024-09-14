
package pipeline

import (
	"go-spider/pipeline/collector"
	"go-spider/pipeline/collector/data"
	"go-spider/spider"
)


type Pipeline interface {
	Start()                          
	Stop()                           
	CollectData(data.DataCell) error 
	CollectFile(data.FileCell) error 
}

func New(sp *spider.Spider) Pipeline {
	return collector.NewCollector(sp)
}
