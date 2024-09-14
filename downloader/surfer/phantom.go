package surfer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	
	
	
	Phantom struct {
		PhantomjsFile string            
		TempJsDir     string            
		jsFileMap     map[string]string 
		CookieJar     *cookiejar.Jar
	}
	
	Response struct {
		Cookies []string
		Body    string
		Error   string
		Header  []struct {
			Name  string
			Value string
		}
	}

	
	Cookie struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Domain string `json:"domain"`
		Path   string `json:"path"`
	}
)

func NewPhantom(phantomjsFile, tempJsDir string, jar ...*cookiejar.Jar) Surfer {
	phantom := &Phantom{
		PhantomjsFile: phantomjsFile,
		TempJsDir:     tempJsDir,
		jsFileMap:     make(map[string]string),
	}
	if len(jar) != 0 {
		phantom.CookieJar = jar[0]
	} else {
		phantom.CookieJar, _ = cookiejar.New(nil)
	}
	if !filepath.IsAbs(phantom.PhantomjsFile) {
		phantom.PhantomjsFile, _ = filepath.Abs(phantom.PhantomjsFile)
	}
	if !filepath.IsAbs(phantom.TempJsDir) {
		phantom.TempJsDir, _ = filepath.Abs(phantom.TempJsDir)
	}
	
	err := os.MkdirAll(phantom.TempJsDir, 0777)
	if err != nil {
		log.Printf("[E] Surfer: %v\n", err)
		return phantom
	}
	phantom.createJsFile("js", js)
	return phantom
}


func (self *Phantom) Download(req Request) (resp *http.Response, err error) {
	var encoding = "utf-8"
	if _, params, err := mime.ParseMediaType(req.GetHeader().Get("Content-Type")); err == nil {
		if cs, ok := params["charset"]; ok {
			encoding = strings.ToLower(strings.TrimSpace(cs))
		}
	}

	req.GetHeader().Del("Content-Type")

	param, err := NewParam(req)
	if err != nil {
		return nil, err
	}

	cookie := ""
	if req.GetEnableCookie() {
		httpCookies := self.CookieJar.Cookies(param.url)
		if len(httpCookies) > 0 {
			surferCookies := make([]*Cookie, len(httpCookies))

			for n, c := range httpCookies {
				surferCookie := &Cookie{Name: c.Name, Value: c.Value, Domain: param.url.Host, Path: "/"}
				surferCookies[n] = surferCookie
			}

			c, err := json.Marshal(surferCookies)
			if err != nil {
				log.Printf("cookie marshal error:%v", err)
			}
			cookie = string(c)
		}
	}

	resp = param.writeback(resp)
	resp.Request.URL = param.url

	var args = []string{
		self.jsFileMap["js"],
		req.GetUrl(),
		cookie,
		encoding,
		param.header.Get("User-Agent"),
		req.GetPostData(),
		strings.ToLower(param.method),
		fmt.Sprint(int(req.GetDialTimeout() / time.Millisecond)),
	}
	if req.GetProxy() != "" {
		args = append([]string{"--proxy=" + req.GetProxy()}, args...)
	}

	for i := 0; i < param.tryTimes; i++ {
		if i != 0 {
			time.Sleep(param.retryPause)
		}

		cmd := exec.Command(self.PhantomjsFile, args...)
		if resp.Body, err = cmd.StdoutPipe(); err != nil {
			continue
		}
		err = cmd.Start()
		if err != nil || resp.Body == nil {
			continue
		}
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		retResp := Response{}
		err = json.Unmarshal(b, &retResp)
		if err != nil {
			continue
		}

		if retResp.Error != "" {
			log.Printf("phantomjs response error:%s", retResp.Error)
			continue
		}

		
		for _, h := range retResp.Header {
			resp.Header.Add(h.Name, h.Value)
		}

		
		for _, c := range retResp.Cookies {
			resp.Header.Add("Set-Cookie", c)
		}
		if req.GetEnableCookie() {
			if rc := resp.Cookies(); len(rc) > 0 {
				self.CookieJar.SetCookies(param.url, rc)
			}
		}
		resp.Body = ioutil.NopCloser(strings.NewReader(retResp.Body))
		break
	}

	if err == nil {
		resp.StatusCode = http.StatusOK
		resp.Status = http.StatusText(http.StatusOK)
	} else {
		resp.StatusCode = http.StatusBadGateway
		resp.Status = err.Error()
	}
	return
}


func (self *Phantom) DestroyJsFiles() {
	p, _ := filepath.Split(self.TempJsDir)
	if p == "" {
		return
	}
	for _, filename := range self.jsFileMap {
		os.Remove(filename)
	}
	if len(WalkDir(p)) == 1 {
		os.Remove(p)
	}
}

func (self *Phantom) createJsFile(fileName, jsCode string) {
	fullFileName := filepath.Join(self.TempJsDir, fileName)
	
	f, _ := os.Create(fullFileName)
	f.Write([]byte(jsCode))
	f.Close()
	self.jsFileMap[fileName] = fullFileName
}