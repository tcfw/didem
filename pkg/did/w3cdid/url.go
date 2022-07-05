package w3cdid

import (
	"net/url"
	"strings"
)

type URL string

func (u URL) Scheme() string {
	return "did"
}

func (u URL) Method() string {
	uri, _ := url.Parse(string(u))
	p := strings.SplitN(uri.Opaque, ":", 2)
	return p[0]
}

func (u URL) Id() string {
	uri, _ := url.Parse(string(u))
	p := strings.SplitN(uri.Opaque, ":", 2)
	if len(p) < 2 {
		return ""
	}

	return p[1]
}

func (u URL) Query() string {
	uri, _ := url.Parse(string(u))
	return uri.RawQuery
}

func (u URL) Fragment() string {
	uri, _ := url.Parse(string(u))
	return uri.Fragment
}
