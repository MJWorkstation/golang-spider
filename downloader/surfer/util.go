package surfer

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"golang.org/x/net/html/charset"
)

func AutoToUTF8(resp *http.Response) error {
	destReader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err == nil {
		resp.Body = &Body{
			ReadCloser: resp.Body,
			Reader:     destReader,
		}
	}
	return err
}


func BodyBytes(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return body, err
}


func UrlEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}


func GetWDPath() string {
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not setted in env.")
	}
	return wd
}


func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}

	panic("util isDirExists not reached")
}


func IsFileExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return !fi.IsDir()
	}

	panic("util isFileExists not reached")
}


func WalkDir(targpath string, suffixes ...string) (dirlist []string) {
	if !filepath.IsAbs(targpath) {
		targpath, _ = filepath.Abs(targpath)
	}
	err := filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		if len(suffixes) == 0 {
			dirlist = append(dirlist, retpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(retpath, suffix) {
				dirlist = append(dirlist, retpath)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("utils.WalkDir: %v\n", err)
		return
	}

	return
}


type Body struct {
	io.ReadCloser
	io.Reader
}

func (b *Body) Read(p []byte) (int, error) {
	return b.Reader.Read(p)
}
