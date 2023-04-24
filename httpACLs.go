package egressproxy

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func (s *HTTPProxy) loadAcls() error {

	log.Printf("Loading ACLs\n")
	hosts, _ := os.LookupEnv("ALLOWED_HOSTS")
	urls, _ := os.LookupEnv("ALLOWED_URLS")

	switch {
	case hosts != "":
		for _, host := range strings.Split(hosts, "\n") {
			log.Printf("compiling host regex: %v\n", host)
			re, err := regexp.Compile(host)
			if err != nil {
				return fmt.Errorf("error compiling %s: %v", host, err)
			}
			s.allowedHosts = append(s.allowedHosts, re)
		}
	case urls != "":
		for _, url := range strings.Split(urls, "\n") {
			log.Printf("compiling url regex: %v\n", url)
			re, err := regexp.Compile(url)
			if err != nil {
				return fmt.Errorf("error compiling %s: %v", url, err)
			}
			s.allowedURLs = append(s.allowedURLs, re)
		}
	}

	log.Printf("Loaded %d hostnames, %d URL regexes\n", len(s.allowedHosts), len(s.allowedURLs))

	return nil
}

func (s *HTTPProxy) checkAccess(r *http.Request) bool {
	switch {
	case check(r.URL.Hostname(), s.allowedHosts):
		log.Printf("Allowing host %v\n", r.URL.Hostname())
		return true
	case check(r.URL.String(), s.allowedURLs):
		log.Printf("Allowing url %v\n", r.RequestURI)
		return true
	default:
		log.Printf("Blocking %v\n", r.URL)
		return false
	}
}

func check(s string, regexes []*regexp.Regexp) (ok bool) {
	for _, re := range regexes {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
