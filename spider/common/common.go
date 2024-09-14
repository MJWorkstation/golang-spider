package common

import (
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)


func CleanHtml(str string, depth int) string {
	if depth > 0 {
		
		re, _ := regexp.Compile("<[\\S\\s]+?>")
		str = re.ReplaceAllStringFunc(str, strings.ToLower)
	}
	if depth > 1 {
		
		re, _ := regexp.Compile("<style[\\S\\s]+?</style>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 2 {
		
		re, _ := regexp.Compile("<script[\\S\\s]+?</script>")
		str = re.ReplaceAllString(str, "")
	}
	if depth > 3 {
		
		re, _ := regexp.Compile("<[\\S\\s]+?>")
		str = re.ReplaceAllString(str, "\n")
	}
	if depth > 4 {
		
		re, _ := regexp.Compile("\\s{2,}")
		str = re.ReplaceAllString(str, "\n")
	}
	return str
}



func ExtractArticle(html string) string {
	
	re := regexp.MustCompile("<[\\S\\s]+?>")
	html = re.ReplaceAllStringFunc(html, strings.ToLower)
	
	re = regexp.MustCompile("<head[\\S\\s]+?</head>")
	html = re.ReplaceAllString(html, "")
	
	re = regexp.MustCompile("<style[\\S\\s]+?</style>")
	html = re.ReplaceAllString(html, "")
	
	re = regexp.MustCompile("<script[\\S\\s]+?</script>")
	html = re.ReplaceAllString(html, "")
	
	re = regexp.MustCompile("<![\\S\\s]+?>")
	html = re.ReplaceAllString(html, "")
	

	
	re = regexp.MustCompile("<[A-Za-z]+[^<]*>([^<>]+)</[A-Za-z]+>")
	ss := re.FindAllStringSubmatch(html, -1)
	

	var maxLen int
	var idx int
	for k, v := range ss {
		l := len([]rune(v[1]))
		if l > maxLen {
			maxLen = l
			idx = k
		}
	}
	

	html = strings.Replace(html, ss[idx][0], `<go-spider id="go-spider">`+ss[idx][1]+`</go-spider>`, -1)
	r := strings.NewReader(html)
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ""
	}
	return dom.Find("go-spider#go-spider").Parent().Text()
}


func Deprive(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, ` `, "", -1)
	return s
}


func Deprive2(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, `\n`, "", -1)
	s = strings.Replace(s, `\r`, "", -1)
	s = strings.Replace(s, `\t`, "", -1)
	s = strings.Replace(s, ` `, "", -1)
	return s
}


func Floor(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f)*pow10_n) / pow10_n
}


func SplitCookies(cookieStr string) (cookies []*http.Cookie) {
	slice := strings.Split(cookieStr, ";")
	for _, v := range slice {
		oneCookie := &http.Cookie{}
		s := strings.Split(v, "=")
		if len(s) == 2 {
			oneCookie.Name = strings.Trim(s[0], " ")
			oneCookie.Value = strings.Trim(s[1], " ")
			cookies = append(cookies, oneCookie)
		}
	}
	return
}

func DecodeString(src, charset string) string {
	return mahonia.NewDecoder(charset).ConvertString(src)
}

func EncodeString(src, charset string) string {
	return mahonia.NewEncoder(charset).ConvertString(src)
}

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}





func GBKToUTF8(src string) string {
	return DecodeString(src, "GB18030")
}


func UnicodeToUTF8(str string) string {
	str = strings.TrimLeft(str, "&#")
	str = strings.TrimRight(str, ";")
	strSlice := strings.Split(str, ";&#")

	for k, s := range strSlice {
		if i, err := strconv.Atoi(s); err == nil {
			strSlice[k] = string(i)
		}
	}
	return strings.Join(strSlice, "")
}


func Unicode16ToUTF8(str string) string {
	i := 0
	if strings.Index(str, `\u`) > 0 {
		i = 1
	}
	strSlice := strings.Split(str, `\u`)
	last := len(strSlice) - 1
	if len(strSlice[last]) > 4 {
		strSlice = append(strSlice, string(strSlice[last][4:]))
		strSlice[last] = string(strSlice[last][:4])
	}
	for ; i <= last; i++ {
		if x, err := strconv.ParseInt(strSlice[i], 16, 32); err == nil {
			strSlice[i] = string(x)
		}
	}
	return strings.Join(strSlice, "")
}



func MakeUrl(path string, schemeAndHost ...string) (string, bool) {
	if string(path[0]) != "/" && strings.ToLower(string(path[0])) != "h" {
		path = "/" + path
	}
	u := path
}