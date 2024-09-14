package app

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"go-spider/crawler"
	"go-spider/distribute"
	"go-spider/pipeline"
	"go-spider/pipeline/collector"
	"go-spider/scheduler"
	"go-spider/spider"
	"go-spider/logs"
	"go-spider/runtime/cache"
	"go-spider/runtime/status"
	"teleport"
)

type (
	App interface {
		Init(mode int, port int, master string, w ...io.Writer) App
		ReInit(mode int, port int, master string, w ...io.Writer) App
		GetAppConf(k ...string) interface{}
		SetAppConf(k string, v interface{}) App
		SpiderPrepare(original []*spider.Spider) App

		Run()
		Stop()
		IsRunning() bool
		IsPause() bool
		IsStopped() bool
		PauseRecover()
		Status() int

		GetSpiderLib() []*spider.Spider
		GetSpiderByName(string) *spider.Spider
		GetSpiderQueue() crawler.SpiderQueue
		GetOutputLib() []string
		GetTaskJar() *distribute.TaskJar
		distribute.Distributer
	}
	Logic struct {
		*cache.AppConf
		*spider.SpiderSpecies
		crawler.SpiderQueue
		*distribute.TaskJar
		crawler.CrawlerPool
		teleport.Teleport
		sum                   [2]uint64
		takeTime              time.Duration
		status                int
		finish                chan bool
		finishOnce            sync.Once
		canSocketLog          bool
		sync.RWMutex
	}
)
var LogicApp = New()

func New() App {
	return newLogic()
}

func newLogic() *Logic {
	return &Logic{
		AppConf:       cache.Task,
		SpiderSpecies: spider.Species,
		status:        status.STOPPED,
		Teleport:      teleport.New(),
		TaskJar:       distribute.NewTaskJar(),
		SpiderQueue:   crawler.NewSpiderQueue(),
		CrawlerPool:   crawler.NewCrawlerPool(),
	}
}

func (self *Logic) GetAppConf(k ...string) interface{} {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	if len(k) == 0 {
		return self.AppConf
	}
	key := strings.Title(k[0])
	acv := reflect.ValueOf(self.AppConf).Elem()
	return acv.FieldByName(key).Interface()
}

func (self *Logic) SetAppConf(k string, v interface{}) App {
	defer func() {
		if err := recover(); err != nil {}
	}()
	if k == "Limit" && v.(int64) <= 0 {
		v = int64(spider.LIMIT)
	} else if k == "DockerCap" && v.(int) < 1 {
		v = int(1)
	}
	acv := reflect.ValueOf(self.AppConf).Elem()
	key := strings.Title(k)
	if acv.FieldByName(key).CanSet() {
		acv.FieldByName(key).Set(reflect.ValueOf(v))
	}
	return self
}

func (self *Logic) Init(mode int, port int, master string, w ...io.Writer) App {
	self.canSocketLog = false
	if len(w) > 0 {}
	self.AppConf.Mode, self.AppConf.Port, self.AppConf.Master = mode, port, master
	self.Teleport = teleport.New()
	self.TaskJar = distribute.NewTaskJar()
	self.SpiderQueue = crawler.NewSpiderQueue()
	self.CrawlerPool = crawler.NewCrawlerPool()

	switch self.AppConf.Mode {
	case status.SERVER:
		if self.checkPort() {
			logs.Log.Informational("当前运行模式为：[Server] 模式")
			self.Teleport.SetAPI(distribute.MasterApi(self)).Server(":" + strconv.Itoa(self.AppConf.Port))
		}

	case status.CLIENT:
		if self.checkAll() {
			logs.Log.Informational("当前运行模式为：[Client] 模式")
			self.Teleport.SetAPI(distribute.SlaveApi(self)).Client(self.AppConf.Master, ":"+strconv.Itoa(self.AppConf.Port))
		}
	case status.OFFLINE:
		logs.Log.EnableStealOne(false)
		return self
	default:
		return self
	}
	return self
}

func (self *Logic) ReInit(mode int, port int, master string, w ...io.Writer) App {
	if !self.IsStopped() {
		self.Stop()
	}
	self.LogRest()
	if self.Teleport != nil {
		self.Teleport.Close()
	}
	if mode == status.UNSET {
		self = newLogic()
		self.AppConf.Mode = status.UNSET
		return self
	}
	self = newLogic().Init(mode, port, master, w...).(*Logic)
	return self
}

func (self *Logic) SpiderPrepare(original []*spider.Spider) App {
	self.SpiderQueue.Reset()
	for _, sp := range original {
		spcopy := sp.Copy()
		spcopy.SetPausetime(self.AppConf.Pausetime)
		if spcopy.GetLimit() == spider.LIMIT {
			spcopy.SetLimit(self.AppConf.Limit)
		} else {
			spcopy.SetLimit(-1 * self.AppConf.Limit)
		}
		self.SpiderQueue.Add(spcopy)
	}
	self.SpiderQueue.AddKeyins(self.AppConf.Keyins)
	return self
}

func (self *Logic) GetOutputLib() []string {
	return collector.DataOutputLib
}

func (self *Logic) GetSpiderLib() []*spider.Spider {
	return self.SpiderSpecies.Get()
}

func (self *Logic) GetSpiderByName(name string) *spider.Spider {
	return self.SpiderSpecies.GetByName(name)
}

func (self *Logic) GetMode() int {
	return self.AppConf.Mode
}

func (self *Logic) GetTaskJar() *distribute.TaskJar {
	return self.TaskJar
}

func (self *Logic) CountNodes() int {
	return self.Teleport.CountNodes()
}

func (self *Logic) GetSpiderQueue() crawler.SpiderQueue {
	return self.SpiderQueue
}

func (self *Logic) Run() {
	if self.AppConf.Mode != status.CLIENT && self.SpiderQueue.Len() == 0 {
		return
	}
	self.finish = make(chan bool)
	self.finishOnce = sync.Once{}
	self.sum[0], self.sum[1] = 0, 0
	self.takeTime = 0
	self.setStatus(status.RUN)
	defer self.setStatus(status.STOPPED)
	switch self.AppConf.Mode {
	case status.OFFLINE:
		self.offline()
	case status.SERVER:
		self.server()
	case status.CLIENT:
		self.client()
	default:
		return
	}
	<-self.finish
}

func (self *Logic) PauseRecover() {
	switch self.Status() {
	case status.PAUSE:
		self.setStatus(status.RUN)
	case status.RUN:
		self.setStatus(status.PAUSE)
	}

	scheduler.PauseRecover()
}

func (self *Logic) Stop() {
	if self.status == status.STOPPED {
		return
	}
	if self.status != status.STOP {
		self.setStatus(status.STOP)
		scheduler.Stop()
		self.CrawlerPool.Stop()
	}
	for !self.IsStopped() {
		time.Sleep(time.Second)
	}
}

func (self *Logic) IsRunning() bool {
	return self.status == status.RUN
}

func (self *Logic) IsPause() bool {
	return self.status == status.PAUSE
}

func (self *Logic) IsStopped() bool {
	return self.status == status.STOPPED
}

func (self *Logic) Status() int {
	self.RWMutex.RLock()
	defer self.RWMutex.RUnlock()
	return self.status
}

func (self *Logic) setStatus(status int) {
	self.RWMutex.Lock()
	defer self.RWMutex.Unlock()
	self.status = status
}

func (self *Logic) offline() {
	self.exec()
}

func (self *Logic) server() {
	defer func() {
		self.finishOnce.Do(func() { close(self.finish) })
	}()
	tasksNum, spidersNum := self.addNewTask()
	if tasksNum == 0 {
		return
	}

	logs.Log.Informational(` *********************************************************************************************************************************** `)
	logs.Log.Informational(" * ")
	logs.Log.Informational(" *                               —— 本次成功添加 %v 条任务，共包含 %v 条采集规则 ——", tasksNum, spidersNum)
	logs.Log.Informational(" * ")
	logs.Log.Informational(` *********************************************************************************************************************************** `)
}

func (self *Logic) addNewTask() (tasksNum, spidersNum int) {
	length := self.SpiderQueue.Len()
	t := distribute.Task{}
	self.setTask(&t)

	for i, sp := range self.SpiderQueue.GetAll() {
		t.Spiders = append(t.Spiders, map[string]string{"name": sp.GetName(), "keyin": sp.GetKeyin()})
		spidersNum++
		if i > 0 && i%10 == 0 && length > 10 {
			one := t
			self.TaskJar.Push(&one)
			tasksNum++
			t.Spiders = []map[string]string{}
		}
	}

	if len(t.Spiders) != 0 {
		one := t
		self.TaskJar.Push(&one)
		tasksNum++
	}
	return
}

func (self *Logic) client() {
	defer func() {
		self.finishOnce.Do(func() { close(self.finish) })
	}()

	for {
		t := self.downTask()
		if self.Status() == status.STOP || self.Status() == status.STOPPED {
			return
		}
		self.taskToRun(t)
		self.sum[0], self.sum[1] = 0, 0
		self.takeTime = 0
		self.exec()
	}
}

func (self *Logic) downTask() *distribute.Task {
ReStartLabel:
	if self.Status() == status.STOP || self.Status() == status.STOPPED {
		return nil
	}
	if self.CountNodes() == 0 && self.TaskJar.Len() == 0 {
		time.Sleep(time.Second)
		goto ReStartLabel
	}

	if self.TaskJar.Len() == 0 {
		self.Request(nil, "task", "")
		for self.TaskJar.Len() == 0 {
			if self.CountNodes() == 0 {
				goto ReStartLabel
			}
			time.Sleep(time.Second)
		}
	}
	return self.TaskJar.Pull()
}

func (self *Logic) taskToRun(t *distribute.Task) {
	self.SpiderQueue.Reset()
	self.setAppConf(t)
	for _, n := range t.Spiders {
		sp := self.GetSpiderByName(n["name"])
		if sp == nil {
			continue
		}
		spcopy := sp.Copy()
		spcopy.SetPausetime(t.Pausetime)
		if spcopy.GetLimit() > 0 {
			spcopy.SetLimit(t.Limit)
		} else {
			spcopy.SetLimit(-1 * t.Limit)
		}
		if v, ok := n["keyin"]; ok {
			spcopy.SetKeyin(v)
		}
		self.SpiderQueue.Add(spcopy)
	}
}

func (self *Logic) exec() {
	count := self.SpiderQueue.Len()
	cache.ResetPageCount()
	pipeline.RefreshOutput()
	scheduler.Init()
	crawlerCap := self.CrawlerPool.Reset(count)
	cache.StartTime = time.Now()
	if self.AppConf.Mode == status.OFFLINE {
		go self.goRun(count)
	} else {
		self.goRun(count)
	}
}

func (self *Logic) goRun(count int) {
	var i int
	for i = 0; i < count && self.Status() != status.STOP; i++ {
	pause:
		if self.IsPause() {
			time.Sleep(time.Second)
			goto pause
		}
		c := self.CrawlerPool.Use()
		if c != nil {
			go func(i int, c crawler.Crawler) {
				c.Init(self.SpiderQueue.GetByIndex(i)).Run()
				self.RWMutex.RLock()
				if self.status != status.STOP {
					self.CrawlerPool.Free(c)
				}
				self.RWMutex.RUnlock()
			}(i, c)
		}
	}
	for ii := 0; ii < i; ii++ {
		s := <-cache.ReportChan
		if (s.DataNum == 0) && (s.FileNum == 0) {
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   无采集结果，用时 %v！\n", s.SpiderName, s.Keyin, s.Time)
			continue
		}
		logs.Log.Informational(" * ")
		switch {
		case s.DataNum > 0 && s.FileNum == 0:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共采集数据 %v 条，用时 %v！\n",
				s.SpiderName, s.Keyin, s.DataNum, s.Time)
		case s.DataNum == 0 && s.FileNum > 0:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共下载文件 %v 个，用时 %v！\n",
				s.SpiderName, s.Keyin, s.FileNum, s.Time)
		default:
			logs.Log.App(" *     [任务小计：%s | KEYIN：%s]   共采集数据 %v 条 + 下载文件 %v 个，用时 %v！\n",
				s.SpiderName, s.Keyin, s.DataNum, s.FileNum, s.Time)
		}
		self.sum[0] += s.DataNum
		self.sum[1] += s.FileNum
	}

	self.takeTime = time.Since(cache.StartTime)
	var prefix = func() string {
		if self.Status() == status.STOP {
			return "任务中途取消："
		}
		return "本次"
	}()
	switch {
	case self.sum[0] > 0 && self.sum[1] == 0:
		logs.Log.App(" *                            —— %s合计采集【数据 %v 条】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[0], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] > 0:
		logs.Log.App(" *                            —— %s合计采集【文件 %v 个】， 实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	case self.sum[0] == 0 && self.sum[1] == 0:
		logs.Log.App(" *                            —— %s无采集结果，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	default:
		logs.Log.App(" *                            —— %s合计采集【数据 %v 条 + 文件 %v 个】，实爬【成功 %v URL + 失败 %v URL = 合计 %v URL】，耗时【%v】 ——",
			prefix, self.sum[0], self.sum[1], cache.GetPageCount(1), cache.GetPageCount(-1), cache.GetPageCount(0), self.takeTime)
	}
	
	if self.AppConf.Mode == status.OFFLINE {
		self.finishOnce.Do(func() { close(self.finish) })
	}
}

func (self *Logic) socketLog() {
	for self.canSocketLog {
		if !ok {
			return
		}
		if self.Teleport.CountNodes() == 0 {
			continue
		}
		self.Teleport.Request(msg, "log", "")
	}
}

func (self *Logic) checkPort() bool {
	if self.AppConf.Port == 0 {
		logs.Log.Warning(" *     —— 分布式端口不能为空")
		return false
	}
	return true
}

func (self *Logic) checkAll() bool {
	if self.AppConf.Master == "" || !self.checkPort() {
		logs.Log.Warning(" *     —— 服务器地址不能为空")
		return false
	}
	return true
}

func (self *Logic) setAppConf(task *distribute.Task) {
	self.AppConf.ThreadNum = task.ThreadNum
	self.AppConf.Pausetime = task.Pausetime
	self.AppConf.OutType = task.OutType
	self.AppConf.DockerCap = task.DockerCap
	self.AppConf.SuccessInherit = task.SuccessInherit
	self.AppConf.FailureInherit = task.FailureInherit
	self.AppConf.Limit = task.Limit
	self.AppConf.ProxyMinute = task.ProxyMinute
	self.AppConf.Keyins = task.Keyins
}
func (self *Logic) setTask(task *distribute.Task) {
	task.ThreadNum = self.AppConf.ThreadNum
	task.Pausetime = self.AppConf.Pausetime
	task.OutType = self.AppConf.OutType
	task.DockerCap = self.AppConf.DockerCap
	task.SuccessInherit = self.AppConf.SuccessInherit
	task.FailureInherit = self.AppConf.FailureInherit
	task.Limit = self.AppConf.Limit
	task.ProxyMinute = self.AppConf.ProxyMinute
	task.Keyins = self.AppConf.Keyins
}
