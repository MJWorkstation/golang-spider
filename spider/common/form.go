package common

import (
	"net/url"
	"strings"
	"go-spider/common/goquery"
	"go-spider/downloader/request"
	"go-spider/spider"
)


type Form struct {
	ctx       *Context
	rule      string
	selection *goquery.Selection
	method    string
	action    string
	fields    url.Values
	buttons   url.Values
}


func NewForm(ctx *Context, rule string, u string, form *goquery.Selection, schemeAndHost ...string) *Form {
	fields, buttons := serializeForm(form)
	if len(schemeAndHost) == 0 {
		aurl, _ := url.Parse(u)
	}
		}