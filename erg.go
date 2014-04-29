package main

import (
  "fmt"
  "flag"
  "net/http"
  "net/url"
  "bufio"
  "sort"
  "os"

  "github.com/xaviershay/grange"
)

var (
  expand bool
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "  usage: erg [opts] QUERY")
		fmt.Fprintln(os.Stderr, "example: erg -e @all")
		fmt.Fprintln(os.Stderr, )

		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr)
	}
	flag.BoolVar(&expand, "e", false, "Do not compress results")
}

func main() {
  flag.Parse()
  var query string
	switch flag.NArg() {
  case 1:
    query = flag.Arg(0)
  default:
		flag.Usage()
		os.Exit(1)
  }
  resp, _ := http.Get("http://localhost:8080/range/list?" +
    url.QueryEscape(query))
  defer resp.Body.Close()
  scanner := bufio.NewScanner(resp.Body)

  result := grange.NewResult()
  for scanner.Scan() {
    result.Add(scanner.Text())
  }
  if result.Cardinality() > 0 {
    if expand {
      strResult := []string{}
      for node := range result.Iter() {
        strResult = append(strResult, node.(string))
      }
      sort.Strings(strResult)
      for _, node := range strResult {
        fmt.Println(node)
      }
    } else {
      fmt.Println(grange.Compress(&result))
    }
  }
}
