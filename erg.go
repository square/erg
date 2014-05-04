package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"

	"github.com/xaviershay/grange"

	goopt "github.com/droundy/goopt"
)

var port = goopt.Int([]string{"-p", "--port"}, 8080, "Port to connect to. Can also be set with RANGE_PORT environment variable.")
var host = goopt.String([]string{"-h", "--host"}, "localhost", "Host to connect to. Can also be set with RANGE_HOST environment variable.")
var expand = goopt.Flag([]string{"-e", "--expand"}, []string{"--no-expand"},
	"Do not compress results", "Compress results (default)")
var noSortResult = goopt.Flag([]string{"--no-sort"}, []string{"-s", "--sort"},
	"Do not sort results. Only relevant with --expand option.", "Sort results (default)")

func main() {
	if envHost := os.Getenv("RANGE_HOST"); len(envHost) > 0 {
		*host = envHost
	}

	if envPort := os.Getenv("RANGE_PORT"); len(envPort) > 0 {
		x, err := strconv.Atoi(envPort)
		if err == nil {
			*port = x
		} else {
			fmt.Fprintf(os.Stderr, "Invalid port in RANGE_PORT: %s\n", envPort)
			os.Exit(1)
		}
	}
	goopt.Parse(nil)

	var query string
	switch len(goopt.Args) {
	case 1:
		query = goopt.Args[0]
	default:
		fmt.Fprintln(os.Stderr, goopt.Usage())
		os.Exit(1)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/range/list?%s",
		*host,
		*port,
		url.QueryEscape(query),
	))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	result := grange.NewResult()
	for scanner.Scan() {
		result.Add(scanner.Text())
	}
	if result.Cardinality() > 0 {
		if *expand {
			strResult := []string{}
			for node := range result.Iter() {
				strResult = append(strResult, node.(string))
			}
			if !*noSortResult {
				sort.Strings(strResult)
			}
			for _, node := range strResult {
				fmt.Println(node)
			}
		} else {
			fmt.Println(grange.Compress(&result))
		}
	}
}
