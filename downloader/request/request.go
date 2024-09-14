package request

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-spider/common/util"
)


type Request struct {
	Spider        string          
	Url           string          
	Rule          string          
	Method        string          
	Header        http.Header     
	EnableCookie  bool            
	PostData      string          
	DialTimeout   time.Duration   
	ConnTimeout   time.Duration   
	TryTimes      int             
	RetryPause    time.Duration   
	RedirectTimes int             
	Temp          Temp            
	TempIsJson    map[string]bool 
	Priority      int             
	Reloadable    bool            
	
	
	
	DownloaderID int

	proxy  string 
	unique string 
	lock   sync.RWMutex
}

const (
	DefaultDialTimeout = 2 * time.Minute 
	DefaultConnTimeout = 2 * time.Minute 
	DefaultTryTimes    = 3               
	DefaultRetryPause  = 2 * time.Second 
)

const (
	SURF_ID    = 0 
	PHANTOM_ID = 1 
)













func (self *Request) Prepare() error {
	
	URL, err := url.Parse(self.Url)
	if err != nil {
		return err
	}
	self.Url = URL.String()

	if self.Method == "" {
		self.Method = "GET"
	} else {
		self.Method = strings.ToUpper(self.Method)
	}

	if self.Header == nil {
		self.Header = make(http.Header)
	}

	if self.DialTimeout < 0 {
		self.DialTimeout = 0
	} else if self.DialTimeout == 0 {
		self.DialTimeout = DefaultDialTimeout
	}

	if self.ConnTimeout < 0 {
		self.ConnTimeout = 0
	} else if self.ConnTimeout == 0 {
		self.ConnTimeout = DefaultConnTimeout
	}

	if self.TryTimes == 0 {
		self.TryTimes = DefaultTryTimes
	}

	if self.RetryPause <= 0 {
		self.RetryPause = DefaultRetryPause
	}

	if self.Priority < 0 {
		self.Priority = 0
	}

	if self.DownloaderID < SURF_ID || self.DownloaderID > PHANTOM_ID {
		self.DownloaderID = SURF_ID
	}

	if self.TempIsJson == nil {
		self.TempIsJson = make(map[string]bool)
	}

	if self.Temp == nil {
		self.Temp = make(Temp)
	}
	return nil
}


func UnSerialize(s string) (*Request, error) {
	req := new(Request)
	return req, json.Unmarshal([]byte(s), req)
}


func (self *Request) Serialize() string {
	for k, v := range self.Temp {
		self.Temp.set(k, v)
		self.TempIsJson[k] = true
	}
	b, _ := json.Marshal(self)
	return strings.Replace(util.Bytes2String(b), `\u0026`, `&`, -1)
}


func (self *Request) Unique() string {
	if self.unique == "" {
		block := md5.Sum([]byte(self.Spider + self.Rule + self.Url + self.Method))
		self.unique = hex.EncodeToString(block[:])
	}
	return self.unique
}


func (self *Request) Copy() *Request {
	reqcopy := new(Request)
	b, _ := json.Marshal(self)
	json.Unmarshal(b, reqcopy)
	return reqcopy
}


func (self *Request) GetUrl() string {
	return self.Url
}


func (self *Request) GetMethod() string {
	return self.Method
}


func (self *Request) SetMethod(method string) *Request {
	self.Method = strings.ToUpper(method)
	return self
}

func (self *Request) SetUrl(url string) *Request {
	self.Url = url
	return self
}

func (self *Request) GetReferer() string {
	return self.Header.Get("Referer")
}

func (self *Request) SetReferer(referer string) *Request {
	self.Header.Set("Referer", referer)
	return self
}

func (self *Request) GetPostData() string {
	return self.PostData
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) SetHeader(key, value string) *Request {
	self.Header.Set(key, value)
	return self
}

func (self *Request) AddHeader(key, value string) *Request {
	self.Header.Add(key, value)
	return self
}

func (self *Request) GetEnableCookie() bool {
	return self.EnableCookie
}

func (self *Request) SetEnableCookie(enableCookie bool) *Request {
	self.EnableCookie = enableCookie
	return self
}

func (self *Request) GetCookies() string {
	return self.Header.Get("Cookie")
}

func (self *Request) SetCookies(cookie string) *Request {
	self.Header.Set("Cookie", cookie)
	return self
}

func (self *Request) GetDialTimeout() time.Duration {
	return self.DialTimeout
}

func (self *Request) GetConnTimeout() time.Duration {
	return self.ConnTimeout
}

func (self *Request) GetTryTimes() int {
	return self.TryTimes
}

func (self *Request) GetRetryPause() time.Duration {
	return self.RetryPause
}

func (self *Request) GetProxy() string {
	return self.proxy
}

func (self *Request) SetProxy(proxy string) *Request {
	self.proxy = proxy
	return self
}

func (self *Request) GetRedirectTimes() int {
	return self.RedirectTimes
}

func (self *Request) GetRuleName() string {
	return self.Rule
}

func (self *Request) SetRuleName(ruleName string) *Request {
	self.Rule = ruleName
	return self
}

func (self *Request) GetSpiderName() string {
	return self.Spider
}

func (self *Request) SetSpiderName(spiderName string) *Request {
	self.Spider = spiderName
	return self
}

func (self *Request) IsReloadable() bool {
	return self.Reloadable
}

func (self *Request) SetReloadable(can bool) *Request {
	self.Reloadable = can
	return self
}



func (self *Request) GetTemp(key string, defaultValue interface{}) interface{} {
	if defaultValue == nil {
		panic("*Request.GetTemp()的defaultValue不能为nil，错误位置：key=" + key)
	}
	self.lock.RLock()
	defer self.lock.RUnlock()

	if self.Temp[key] == nil {
		return defaultValue
	}

	if self.TempIsJson[key] {
		return self.Temp.get(key, defaultValue)
	}

	return self.Temp[key]
}

func (self *Request) GetTemps() Temp {
	return self.Temp
}

func (self *Request) SetTemp(key string, value interface{}) *Request {
	self.lock.Lock()
	self.Temp[key] = value
	delete(self.TempIsJson, key)
	self.lock.Unlock()
	return self
}

func (self *Request) SetTemps(temp map[string]interface{}) *Request {
	self.lock.Lock()
	self.Temp = temp
	self.TempIsJson = make(map[string]bool)
	self.lock.Unlock()
	return self
}

func (self *Request) GetPriority() int {
	return self.Priority
}

func (self *Request) SetPriority(priority int) *Request {
	self.Priority = priority
	return self
}

func (self *Request) GetDownloaderID() int {
	return self.DownloaderID
}

func (self *Request) SetDownloaderID(id int) *Request {
	self.DownloaderID = id
	return self
}

func (self *Request) MarshalJSON() ([]byte, error) {
	for k, v := range self.Temp {
		if self.TempIsJson[k] {
			continue
		}
		self.Temp.set(k, v)
		self.TempIsJson[k] = true
	}
	b, err := json.Marshal(*self)
	return b, err
}
