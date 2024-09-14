package downloader

import (
	"go-spider/downloader/request"
	"go-spider/spider"
)

type Downloader interface {
	Download(*spider.Spider, *request.Request) *spider.Context
}
