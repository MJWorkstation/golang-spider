package collector

import (
	"encoding/csv"
	"fmt"
	"os"
	"go-spider/common/util"
	"go-spider/config"
	"go-spider/logs"
	"go-spider/runtime/cache"
)

func init() {
	DataOutput["csv"] = func(self *Collector) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("%v", p)
			}
		}()
		var (
			namespace = util.FileNameReplace(self.namespace())
			sheets    = make(map[string]*csv.Writer)
		)
		for _, datacell := range self.dataDocker {
			var subNamespace = util.FileNameReplace(self.subNamespace(datacell))
			if _, ok := sheets[subNamespace]; !ok {
				folder := config.TEXT_DIR + "/" + cache.StartTime.Format("2006-01-02 150405") + "/" + joinNamespaces(namespace, subNamespace)
				filename := fmt.Sprintf("%v/%v-%v.csv", folder, self.sum[0], self.sum[1])

				
				f, err := os.Stat(folder)
				if err != nil || !f.IsDir() {
					if err := os.MkdirAll(folder, 0777); err != nil {
						logs.Log.Error("Error: %v\n", err)
					}
				}

				
				file, err := os.Create(filename)

				if err != nil {
					logs.Log.Error("%v", err)
					continue
				}
				defer func() {
					
					sheets[subNamespace].Flush()
					
					file.Close()
				}()

				file.WriteString("\xEF\xBB\xBF") 

				sheets[subNamespace] = csv.NewWriter(file)
				th := self.MustGetRule(datacell["RuleName"].(string)).ItemFields
				if self.Spider.OutDefaultField() {
					th = append(th, "当前链接", "上级链接", "下载时间")
				}
				sheets[subNamespace].Write(th)
			}

			row := []string{}
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					row = append(row, v)
				} else {
					row = append(row, util.JsonString(vd[title]))
				}
			}
			if self.Spider.OutDefaultField() {
				row = append(row, datacell["Url"].(string))
				row = append(row, datacell["ParentUrl"].(string))
				row = append(row, datacell["DownloadTime"].(string))
			}
			sheets[subNamespace].Write(row)
		}
		return
	}
}
