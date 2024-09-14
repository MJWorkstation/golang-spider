package surfer

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
	"goutil"
	"go-spider/downloader/surfer/agent"
)


type Surf struct {
	CookieJar *cookiejar.Jar
}


func New(jar ...*cookiejar.Jar) Surfer {
	s := new(Surf)
	if len(jar) != 0 {
		s.CookieJar = jar[0]
	} else {
		s.CookieJar, _ = cookiejar.New(nil)
	}
	return s
}


func (self *Surf) Download(req Request) (resp *http.Response, err error) {
	param, err := NewParam(req)
	if err != nil {
		return nil, err
	}
	param.header.Set("Connection", "close")
	param.client = self.buildClient(param)
	resp, err = self.httpRequest(param)

	if err == nil {
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			var gzipReader *gzip.Reader
			gzipReader, err = gzip.NewReader(resp.Body)
			if err == nil {
				resp.Body = gzipReader
			}

		case "deflate":
			resp.Body = flate.NewReader(resp.Body)

		case "zlib":
			var readCloser io.ReadCloser
			readCloser, err = zlib.NewReader(resp.Body)
			if err == nil {
				resp.Body = readCloser
			}
		}
	}

	resp = param.writeback(resp)

	return
}

var dnsCache = &DnsCache{ipPortLib: goutil.AtomicMap()}


type DnsCache struct {
	ipPortLib goutil.Map
}


func (d *DnsCache) Reg(addr, ipPort string) {
	d.ipPortLib.Store(addr, ipPort)
}


func (d *DnsCache) Del(addr string) {
	d.ipPortLib.Delete(addr)
}


func (d *DnsCache) Query(addr string) (string, bool) {
	ipPort, ok := d.ipPortLib.Load(addr)
	if !ok {
		return "", false
	}
	return ipPort.(string), true
}


func (self *Surf) buildClient(param *Param) *http.Client {
	client := &http.Client{
		CheckRedirect: param.checkRedirect,
	}

	if param.enableCookie {
		client.Jar = self.CookieJar
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			var (
				c          net.Conn
				err        error
				ipPort, ok = dnsCache.Query(addr)
			)
			if !ok {
				ipPort = addr
				defer func() {
					if err == nil {
						dnsCache.Reg(addr, c.RemoteAddr().String())
					}
				}()
			} else {
				defer func() {
					if err != nil {
						dnsCache.Del(addr)
					}
				}()
			}
			c, err = net.DialTimeout(network, ipPort, param.dialTimeout)
			if err != nil {
				return nil, err
			}
			if param.connTimeout > 0 {
				c.SetDeadline(time.Now().Add(param.connTimeout))
			}
			return c, nil
		},
	}

	if param.proxy != nil {
		transport.Proxy = http.ProxyURL(param.proxy)
	}

	if strings.ToLower(param.url.Scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport
	return client
}


func (self *Surf) httpRequest(param *Param) (resp *http.Response, err error) {
	req, err := http.NewRequest(param.method, param.url.String(), param.body)
	if err != nil {
		return nil, err
	}

	req.Header = param.header

	if param.tryTimes <= 0 {
		for {
			resp, err = param.client.Do(req)
			if err != nil {
				if !param.enableCookie {
					l := len(agent.UserAgents["common"])
					r := rand.New(rand.NewSource(time.Now().UnixNano()))
					req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
				}
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	} else {
		for i := 0; i < param.tryTimes; i++ {
			resp, err = param.client.Do(req)
			if err != nil {
				if !param.enableCookie {
					l := len(agent.UserAgents["common"])
					r := rand.New(rand.NewSource(time.Now().UnixNano()))
					req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
				}
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	}

	return resp, err
}
