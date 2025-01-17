package agent

import (
	"bytes"
	"math/rand"
	"runtime"
	"strings"
	"text/template"
	"time"
)


type TemplateData struct {
	Name string
	Ver  string
	OSN  string
	OSV  string
	Coms string
}


type OSAttributes struct {
	
	OSName string
	
	OSVersion string
	
	Comments []string
}

const (
	
	Windows int = iota
	
	Linux
	
	Macintosh
)


var DefaultOSAttributes = map[int]OSAttributes{
	Windows:   {"Windows NT", "6.3", []string{"x64"}},
	Linux:     {"Linux", "3.16.1", []string{"x64"}},
	Macintosh: {"Intel Mac OS X", "10_6_8", []string{}},
}

type (
	
	Formats map[string]string

	
	UAData struct {
		TopVersion string
		DefaultOS  int
		Formats    Formats
	}

	UATable map[string]UAData
)


var Database = UATable{
	"chrome": {
		"37.0.2049.0",
		Formats{
			"37": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"36": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"35": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"34": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"33": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"32": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"31": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
			"30": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) Chrome/{{.Ver}} Safari/537.36",
		},
	},
	"firefox": {
		"31.0",
		Formats{
			"31": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:31.0) Gecko/20100101 Firefox/{{.Ver}}",
			"30": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:30.0) Gecko/20120101 Firefox/{{.Ver}}",
			"29": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:29.0) Gecko/20120101 Firefox/{{.Ver}}",
			"28": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:28.0) Gecko/20100101 Firefox/{{.Ver}}",
			"27": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:27.0) Gecko/20130101 Firefox/{{.Ver}}",
			"26": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:26.0) Gecko/20121011 Firefox/{{.Ver}}",
			"25": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:25.0) Gecko/20100101 Firefox/{{.Ver}}",
		},
	},
	"msie": {
		"10.0",
		Formats{
			"10": "Mozilla/5.0 (compatible; MSIE 10.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/5.0; .NET CLR 3.5.30729)",
			"9":  "Mozilla/5.0 (compatible; MSIE 9.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/5.0; .NET CLR 3.0.30729)",
			"8":  "Mozilla/5.0 (compatible; MSIE 8.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}Trident/4.0; .NET CLR 3.0.04320)",
			"7":  "Mozilla/4.0 (compatible; MSIE 7.0; {{.OSN}} {{.OSV}}{{if .Coms}}{{.Coms}}; {{end}}.NET CLR 2.0.50727)",
		},
	},
	"opera": {
		"12.14",
		Formats{
			"12": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.9.181 Version/{{.Ver}}",
			"11": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.7.62 Version/{{.Ver}}",
			"10": "Opera/9.80 ({{.OSN}} {{.OSV}}; U{{.Coms}}) Presto/2.2.15 Version/{{.Ver}}",
			"9":  "Opera/9.00 ({{.OSN}} {{.OSV}}; U{{.Coms}})",
		},
	},
	"safari": {
		"6.0",
		Macintosh,
		Formats{
			"6": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/536.26 (KHTML, like Gecko) Version/{{.Ver}} Safari/8536.25",
			"5": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/531.2+ (KHTML, like Gecko) Version/{{.Ver}} Safari/531.2+",
			"4": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/528.16 (KHTML, like Gecko) Version/{{.Ver}} Safari/528.16",
		},
	},
	"itunes": {
		"9.1.1",
		Macintosh,
		Formats{
			"9": "iTunes/{{.Ver}}",
			"8": "iTunes/{{.Ver}}",
			"7": "iTunes/{{.Ver}} (Macintosh; U; PPC Mac OS X 10.4.7{{.Coms}})",
			"6": "iTunes/{{.Ver}} (Macintosh; U; PPC Mac OS X 10.4.5{{.Coms}})",
		},
	},
	"aol": {
		"9.7",
		Formats{
			"9": "Mozilla/5.0 (compatible; MSIE 9.0; AOL {{.Ver}}; AOLBuild 4343.19; {{.OSN}} {{.OSV}}; WOW64; Trident/5.0; FunWebProducts{{.Coms}})",
			"8": "Mozilla/4.0 (compatible; MSIE 7.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}; GTB5; .NET CLR 1.1.4322; .NET CLR 2.0.50727{{.Coms}})",
			"7": "Mozilla/4.0 (compatible; MSIE 7.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}; FunWebProducts{{.Coms}})",
			"6": "Mozilla/4.0 (compatible; MSIE 6.0; AOL {{.Ver}}; {{.OSN}} {{.OSV}}{{.Coms}})",
		},
	},
	"konqueror": {
		"4.9",
		Linux,
		Formats{
			"4": "Mozilla/5.0 (compatible; Konqueror/4.0; {{.OSN}}{{.Coms}}) KHTML/4.0.3 (like Gecko)",
			"3": "Mozilla/5.0 (compatible; Konqueror/3.0-rc6; i686 {{.OSN}}; 20021127{{.Coms}})",
			"2": "Mozilla/5.0 (compatible; Konqueror/2.1.1; {{.OSN}}{{.Coms}})",
		},
	},
	"netscape": {
		"9.1.0285",
		
		Formats{
			"9": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.9.2.4{{.Coms}}) Gecko/20070321 Netscape/{{.Ver}}",
			"8": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.7.5{{.Coms}}) Gecko/20050519 Netscape/{{.Ver}}",
			"7": "Mozilla/5.0 ({{.OSN}}; U; {{.OSN}} {{.OSV}}; rv:1.0.1{{.Coms}}) Gecko/20020921 Netscape/{{.Ver}}",
		},
	},
	"lynx": {
		"2.8.8dev.3",
		Linux,
		Formats{
			"2": "Lynx/{{.Ver}} libwww-FM/2.14 SSL-MM/1.4.1",
			"1": "Lynx (textmode)",
		},
	},
	"googlebot": {
		"2.1",
		Linux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Googlebot/{{.Ver}}; +http:
			"1": "Googlebot/{{.Ver}} (+http:
		},
	},
	"bingbot": {
		"2.0",
		
		Formats{
			"2": "Mozilla/5.0 (compatible; bingbot/{{.Ver}}; +http:
		},
	},
	"yahoobot": {
		"2.0",
		Linux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Yahoo! Slurp; http:
		},
	},
	"default": {
		"1.0",
		Linux,
		Formats{
			"1": "{{.Name}}/{{.Ver}} ({{.OSN}} {{.OSV}}{{.Coms}})",
		},
	},
}


var UserAgents = map[string][]string{}

func init() {
	for browser, userAgentData := range Database {
		if browser == "default" {
			continue
		}
		os := userAgentData.DefaultOS
		osAttribs := DefaultOSAttributes[os]
		for version, _ := range userAgentData.Formats {
			ua := createFromDetails(
				browser,
				version,
				osAttribs.OSName,
				osAttribs.OSVersion,
				osAttribs.Comments)
			UserAgents["all"] = append(UserAgents["all"], ua)

			if browser != "itunes" && browser != "lynx" && browser != "googlebot" && browser != "bingbot" && browser != "yahoobot" {
				UserAgents["common"] = append(UserAgents["common"], ua)
			}
		}
	}
	l := len(UserAgents["common"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(l)
	UserAgents["all"][0], UserAgents["all"][idx] = UserAgents["all"][idx], UserAgents["all"][0]
	UserAgents["common"][0], UserAgents["common"][idx] = UserAgents["common"][idx], UserAgents["common"][0]
}


func CreateReal() string {
	return createFromDetails("Surfer", "1.0", osName(), osVersion(), []string{runtime.Version()})
}


func CreateDefault(browser string) string {
	bn := strings.ToLower(browser)
	data := Database[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		data.TopVersion,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}


func CreateVersion(browser, version string) string {
	bn := strings.ToLower(browser)
	data := Database[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		version,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}


func TopVersion(bname string) string {
	bname = strings.ToLower(bname)
	data, ok := Database[bname]
	if ok {
		return data.TopVersion
	}
	return Database["default"].TopVersion
}

func Format(bname, bver string) string {
	bname = strings.ToLower(bname)
	majVer := strings.Split(bver, ".")[0]
	data, ok := Database[bname]
	if ok {
		format, ok := data.Formats[majVer]
		if ok {
			return format
		} else {
			top := TopVersion(bname)
			majVer = strings.Split(top, ".")[0]
			return data.Formats[majVer]
		}
	}

	return Database["default"].Formats["1"]
}


func createFromDetails(bname, bver, osname, osver string, c []string) string {
	if bver == "" {
		bver = TopVersion(bname)
	}
	comments := strings.Join(c, "; ")
	if comments != "" {
		comments = "; " + comments
	}

	data := TemplateData{bname, bver, osname, osver, comments}
	buff := &bytes.Buffer{}
	t := template.New("formatter")
	t.Parse(Format(bname, bver))
	t.Execute(buff, data)

	return buff.String()
}
