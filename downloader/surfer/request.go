package surfer

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	Request interface {
		
		GetUrl() string
		
		GetMethod() string
		
		GetPostData() string
		
		GetHeader() http.Header
		
		GetEnableCookie() bool
		
		GetDialTimeout() time.Duration
		
		GetConnTimeout() time.Duration
		
		GetTryTimes() int
		
		GetRetryPause() time.Duration
		
		GetProxy() string
		
		GetRedirectTimes() int
		
		GetDownloaderID() int
	}

	
	DefaultRequest struct {
		
		Url string
		
		Method string
		
		Header http.Header
		
		EnableCookie bool
		
		PostData string
		
		DialTimeout time.Duration
		
		ConnTimeout time.Duration
		
		TryTimes int
		
		RetryPause time.Duration
		
		RedirectTimes int
		
		Proxy string

		DownloaderID int

		once sync.Once
	}
)

const (
	SurfID             = 0               
	PhomtomJsID        = 1               
	DefaultMethod      = "GET"           
	DefaultDialTimeout = 2 * time.Minute 
	DefaultConnTimeout = 2 * time.Minute 
	DefaultTryTimes    = 3               
	DefaultRetryPause  = 2 * time.Second 
)

func (self *DefaultRequest) prepare() {
	if self.Method == "" {
		self.Method = DefaultMethod
	}
	self.Method = strings.ToUpper(self.Method)

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

	if self.DownloaderID != PhomtomJsID {
		self.DownloaderID = SurfID
	}
}


func (self *DefaultRequest) GetUrl() string {
	self.once.Do(self.prepare)
	return self.Url
}


func (self *DefaultRequest) GetMethod() string {
	self.once.Do(self.prepare)
	return self.Method
}


func (self *DefaultRequest) GetPostData() string {
	self.once.Do(self.prepare)
	return self.PostData
}


func (self *DefaultRequest) GetHeader() http.Header {
	self.once.Do(self.prepare)
	return self.Header
}


func (self *DefaultRequest) GetEnableCookie() bool {
	self.once.Do(self.prepare)
	return self.EnableCookie
}


func (self *DefaultRequest) GetDialTimeout() time.Duration {
	self.once.Do(self.prepare)
	return self.DialTimeout
}


func (self *DefaultRequest) GetConnTimeout() time.Duration {
	self.once.Do(self.prepare)
	return self.ConnTimeout
}


func (self *DefaultRequest) GetTryTimes() int {
	self.once.Do(self.prepare)
	return self.TryTimes
}


func (self *DefaultRequest) GetRetryPause() time.Duration {
	self.once.Do(self.prepare)
	return self.RetryPause
}


func (self *DefaultRequest) GetProxy() string {
	self.once.Do(self.prepare)
	return self.Proxy
}


func (self *DefaultRequest) GetRedirectTimes() int {
	self.once.Do(self.prepare)
	return self.RedirectTimes
}


func (self *DefaultRequest) GetDownloaderID() int {
	self.once.Do(self.prepare)
	return self.DownloaderID
}
