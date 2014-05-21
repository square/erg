package erg

import (
	"bufio"
	"fmt"
	"github.com/xaviershay/grange"
	"net/http"
	"net/url"
	"sort"
)

// Erg type
// Sort boolean - turn it off/on for sorting on expand
// default is true
type Erg struct {
	host string
	port int
	Sort bool
}

// New(address string) returns a new erg
// takes two arguments
// host - hostname default - range
// port - port default - 80
func New(host string, port int) *Erg {
	return &Erg{host: host, port: port, Sort: true}
}

// Expand takes a range expression as argument
// and returns an slice of strings as result
// err is set to nil on success
func (e *Erg) Expand(query string) (result []string, err error) {

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/range/list?%s",
		e.host,
		e.port,
		url.QueryEscape(query),
	))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
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
			sort.Strings(result)
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
