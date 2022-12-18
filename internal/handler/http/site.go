package http

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
)

//	Content info about ever site
//
type Site struct {
	DomainName  *regexp.Regexp
	Certificate *tls.Certificate
	Paths       []*Path
}

func NewSite(domainName string, config map[string]interface{}) (*Site, error) {
	var (
		certificate tls.Certificate
	)
	if domainName == "" || domainName == "*" {
		domainName = ".*"
	}
	domain, err := regexp.Compile(domainName)
	if err != nil {
		return nil, err
	}
	paths := make([]*Path, 0)
	if config["certificate"] != nil && config["certificateKey"] != nil {
		certStr, ok := config["certificate"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler certificate %v", config["certificate"])
		}
		certKeyStr, ok := config["certificateKey"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler certificate key %v", config["certificateKey"])
		}
		certificate, err = tls.LoadX509KeyPair(certStr, certKeyStr)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

	}
	if config["servers"] != nil {
		p, err := NewPath("/", config)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p)
	} else {
		for k, v := range config {
			if strings.Contains(k, "/") {
				conf := v.(map[string]interface{})
				p, err := NewPath(k, conf)
				if err != nil {
					return nil, err
				}
				paths = append(paths, p)

			}

		}
	}

	return &Site{
		DomainName:  domain,
		Certificate: &certificate,
		Paths:       paths,
	}, nil

}
