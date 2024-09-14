package surfer

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
)

var (
	surf         Surfer
	phantom      Surfer
	once_surf    sync.Once
	once_phantom sync.Once
	tempJsDir    = "./tmp"
	
	phantomjsFile = `./phantomjs`
	cookieJar, _  = cookiejar.New(nil)
)

func Download(req Request) (resp *http.Response, err error) {
	switch req.GetDownloaderID() {
	case SurfID:
		once_surf.Do(func() { surf = New(cookieJar) })
		resp, err = surf.Download(req)
	case PhomtomJsID:
		once_phantom.Do(func() { phantom = NewPhantom(phantomjsFile, tempJsDir, cookieJar) })
		resp, err = phantom.Download(req)
	}
	return
}


func DestroyJsFiles() {
	if pt, ok := phantom.(*Phantom); ok {
		pt.DestroyJsFiles()
	}
}


type Surfer interface {
	Download(Request) (resp *http.Response, err error)
}
