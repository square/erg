package erg

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"

	"github.com/fvbommel/sortorder"
	"github.com/square/grange"
)

// Erg type
// Sort boolean - turn it off/on for sorting on expand
// default is true
type Erg struct {
	host   string
	port   int
	ssl    bool
	Sort   bool
	client *http.Client
}

// New(address string) returns a new erg
// takes two arguments
// host - hostname default - localhost
// port - port default - 8080
// ssl - use https or not default - false
func New(host string, port int) *Erg {
	// TODO: Remove this with go 1.4
	// http://stackoverflow.com/questions/25008571/golang-issue-x509-cannot-verify-signature-algorithm-unimplemented-on-net-http
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion:               tls.VersionTLS11,
			PreferServerCipherSuites: true,
		},
	}
	client := &http.Client{Transport: tr}
	return &Erg{
		host:   host,
		port:   port,
		ssl:    false,
		Sort:   true,
		client: client,
	}
}

func NewWithSsl(host string, port int) *Erg {
	e := New(host, port)
	e.ssl = true
	return e
}

func NewWithClient(client *http.Client, host string, port int, ssl bool) *Erg {
	return &Erg{
		host:   host,
		port:   port,
		ssl:    ssl,
		Sort:   true,
		client: client,
	}
}

// Expand takes a range expression as argument
// and returns an slice of strings as result
// err is set to nil on success
func (e *Erg) Expand(query string) (result []string, err error) {
	protocol := "http"

	if e.ssl {
		protocol = "https"
	}

	resp, err := e.client.Get(fmt.Sprintf("%s://%s:%d/range/list?%s",
		protocol,
		e.host,
		e.port,
		url.QueryEscape(query),
	))
	if err != nil {
		return nil, err
	}

	// "When err is nil, resp always contains a non-nil resp.Body. Caller should
	// close resp.Body when done reading from it." https://golang.org/pkg/net/http/#Client.Get
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}
		return nil, errors.New(string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	grangeResult := grange.NewResult()
	for scanner.Scan() {
		grangeResult.Add(scanner.Text())
	}

	if grangeResult.Cardinality() > 0 {
		for node := range grangeResult.Iter() {
			result = append(result, node.(string))
		}
		if e.Sort {
			sort.Sort(sortorder.Natural(result))
		}
	}

	return result, nil
}

// Compress takes a slice of strings as argument
// and returns a compressed form.
func (*Erg) Compress(nodes []string) (result string) {
	grangeResult := grange.NewResult()
	for _, node := range nodes {
		grangeResult.Add(node)
	}
	return grange.Compress(&grangeResult)
}
