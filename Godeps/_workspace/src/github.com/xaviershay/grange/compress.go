package grange

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"vbom.ml/util/sortorder"
)

var (
	_ = fmt.Println
)

// Normalizes a result set into a minimal range expression, such as
// +{foo,bar}.example.com+.
func Compress(nodes *Result) string {
	noDomain := []string{}
	domains := map[string][]string{}
	for node := range nodes.Iter() {
		tokens := strings.SplitN(node.(string), ".", 2)
		if len(tokens) == 2 {
			domains[tokens[1]] = append(domains[tokens[1]], tokens[0])
		} else {
			noDomain = append(noDomain, node.(string))
		}
	}
	sort.Sort(sortorder.Natural(noDomain))

	result := compressNumeric(noDomain)
	var domainKeys = []string{}
	for domain, _ := range domains {
		domainKeys = append(domainKeys, domain)
	}
	sort.Sort(sortorder.Natural(domainKeys))

	for _, domain := range domainKeys {
		domainNodes := domains[domain]
		sort.Sort(sortorder.Natural(domainNodes))
		domainNodes = compressNumeric(domainNodes)
		joined := strings.Join(domainNodes, ",")
		if len(domainNodes) > 1 {
			joined = "{" + joined + "}"
		}
		result = append(result, joined+"."+domain)
	}
	return strings.Join(result, ",")
}

func numericExpansionFor(prefix string, start string, end string, suffix string) string {
	endN, _ := strconv.Atoi(end)
	startN, _ := strconv.Atoi(start)

	if startN == endN {
		return fmt.Sprintf("%s%s%s", prefix, end, suffix)
	} else {
		return fmt.Sprintf("%s%s..%d%s", prefix, start, endN, suffix)
	}
}

func compressNumeric(nodes []string) []string {
	r := regexp.MustCompile("^(.*?)(\\d+)([^\\d]*)$")

	result := []string{}
	currentPrefix := ""
	currentSuffix := ""
	currentNstr := ""
	start := ""
	startN := -1
	currentN := -1

	flush := func() {
		if startN > -1 {
			result = append(result, numericExpansionFor(currentPrefix, start, currentNstr, currentSuffix))
			startN = -1
			currentPrefix = ""
			currentSuffix = ""
			currentN = -1
			currentNstr = ""
		}
	}

	for _, node := range nodes {
		match := r.FindStringSubmatch(node)

		if match == nil {
			flush()
			result = append(result, node)
		} else {
			prefix := match[1]
			n := match[2]
			suffix := match[3]

			if prefix != currentPrefix || suffix != currentSuffix {
				flush()
			}

			//if len(n) != len(currentNstr) {
			//flush
			newN, _ := strconv.Atoi(n)

			if zeroCount(n) != zeroCount(currentNstr) && len(n) != len(currentNstr) {
				flush()
			}

			if startN < 0 || newN != currentN+1 {
				// first in run
				flush()
				start = n
				startN = newN
			}

			currentNstr = n
			currentN = newN
			currentPrefix = prefix
			currentSuffix = suffix
		}
	}
	flush()
	return result
}

func zeroCount(n string) int {
	count := 0
	for _, c := range n {
		if c != '0' {
			break
		}
		count++
	}
	return count
}
