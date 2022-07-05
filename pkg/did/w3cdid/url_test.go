package w3cdid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type URLParts struct {
	Scheme   string
	Method   string
	Id       string
	Query    string
	Fragment string
}

func TestURLDecodes(t *testing.T) {
	//Very basic tests, mostly silly - the net/url pkg is very extensive in test cases
	tests := map[string]URLParts{
		"did":                          {Scheme: "did", Method: "", Id: "", Query: "", Fragment: ""},
		"did:":                         {Scheme: "did", Method: "", Id: "", Query: "", Fragment: ""},
		"did:example":                  {Scheme: "did", Method: "example", Id: "", Query: "", Fragment: ""},
		"did:example:":                 {Scheme: "did", Method: "example", Id: "", Query: "", Fragment: ""},
		"did:example:1234":             {Scheme: "did", Method: "example", Id: "1234", Query: "", Fragment: ""},
		"did:example:1234?h=1":         {Scheme: "did", Method: "example", Id: "1234", Query: "h=1", Fragment: ""},
		"did:example:1234?h=1#b1":      {Scheme: "did", Method: "example", Id: "1234", Query: "h=1", Fragment: "b1"},
		"did:example:1234:abc?h=1#b1":  {Scheme: "did", Method: "example", Id: "1234:abc", Query: "h=1", Fragment: "b1"},
		"did:example:?h=1#b1":          {Scheme: "did", Method: "example", Id: "", Query: "h=1", Fragment: "b1"},
		"did:example:1234/key1?h=1#b1": {Scheme: "did", Method: "example", Id: "1234/key1", Query: "h=1", Fragment: "b1"},
		"did:abc:1234/key1?h=1#b1":     {Scheme: "did", Method: "abc", Id: "1234/key1", Query: "h=1", Fragment: "b1"},
		"did:abc:/key1?h=1#b1":         {Scheme: "did", Method: "abc", Id: "/key1", Query: "h=1", Fragment: "b1"},
		"aaa:bbb:ccc?h=1#b1":           {Scheme: "did", Method: "bbb", Id: "ccc", Query: "h=1", Fragment: "b1"},
		"aaa:bbb:ccc?h=1?b=1":          {Scheme: "did", Method: "bbb", Id: "ccc", Query: "h=1?b=1", Fragment: ""},
		"aaa:bbb:ccc?h=1&b=1":          {Scheme: "did", Method: "bbb", Id: "ccc", Query: "h=1&b=1", Fragment: ""},
		"aaa:bbb:ccc??b=1":             {Scheme: "did", Method: "bbb", Id: "ccc", Query: "?b=1", Fragment: ""},
		"aaa:bbb:ccc##b=1":             {Scheme: "did", Method: "bbb", Id: "ccc", Query: "", Fragment: "#b=1"},
	}

	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			tk := URL(k)
			Scheme := tk.Scheme()
			Method := tk.Method()
			Id := tk.Id()
			Query := tk.Query()
			Fragment := tk.Fragment()

			assert.Equal(t, test.Scheme, Scheme)
			assert.Equal(t, test.Method, Method)
			assert.Equal(t, test.Id, Id)
			assert.Equal(t, test.Query, Query)
			assert.Equal(t, test.Fragment, Fragment)
		})
	}
}
