package pipeline

import (
	"sort"
	"go-spider/pipeline/collector"
	"go-spider/common/kafka"
	"go-spider/common/mgo"
	"go-spider/common/mysql"
	"go-spider/runtime/cache"
)


func init() {
	for out, _ := range collector.DataOutput {
		collector.DataOutputLib = append(collector.DataOutputLib, out)
	}
	sort.Strings(collector.DataOutputLib)
}


func RefreshOutput() {
	switch cache.Task.OutType {
	case "mgo":
		mgo.Refresh()
	case "mysql":
		mysql.Refresh()
	case "kafka":
		kafka.Refresh()
	}
}
