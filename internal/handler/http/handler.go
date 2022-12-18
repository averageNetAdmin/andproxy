package http

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

//	Separate to sites - virtula hosts
//
type Handler struct {
	Secure bool
	Port   string
	Sites  []*Site
	logger *log.Logger
}

//	Create handler from yaml file
//
func NewHandler(configPath, port string, secure bool) (*Handler, error) {
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	configBytes = []byte(strings.ToLower(string(configBytes)))
	config := make(map[string]interface{}, 0)
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	logDir, ok := config["logdir"].(string)
	if !ok {
		logDir = fmt.Sprintf("/var/log/andproxy/http_%s/", port)
	}

	var sites []*Site
	sitesS, ok := config["sites"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no sites avaible")
	}
	for k, v := range sitesS {
		conf := v.(map[string]interface{})
		s, err := NewSite(k, conf)
		if err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}

	if config["secure"] != nil && !secure {
		secure, ok = config["secure"].(bool)
		if !ok {
			return nil, fmt.Errorf("invalid handler secure level %v", config["secure"])
		}
	}

	err = os.MkdirAll(logDir, 0644)
	if err != nil {
		return nil, err
	}
	logFile := fmt.Sprintf("%s/http_%s.log", logDir, port)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, " ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)

	for _, v := range sites {
		fmt.Println(v)
	}

	return &Handler{
		Sites:  sites,
		Port:   port,
		Secure: secure,
		logger: logger,
	}, err

}

//	Run handler job gorutine
//
func (s *Handler) Listen() {
	go s.listen()
}

//	If conn must be secure - check cert
//	Get requests and delegate they to handle function
//
func (s *Handler) listen() {

	if s.Secure {
		certs := make([]tls.Certificate, 0)
		for i := 0; i < len(s.Sites); i++ {
			certs = append(certs, *s.Sites[i].Certificate)
		}

		TLSConf := &tls.Config{
			Certificates: certs,
		}

		server := &http.Server{
			Addr:      fmt.Sprintf(":%s", s.Port),
			TLSConfig: TLSConf,
			Handler:   s,
		}

		server.ListenAndServeTLS("", "")
	} else {
		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", s.Port),
			Handler: s,
		}
		server.ListenAndServe()
	}
}

//	handler for http.Server
//
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	//	Filter requset by site domain name
	reqSite, _, err := net.SplitHostPort(r.URL.Host)
	if err == nil {
		fmt.Println(err)
	}
	var s *Site
	for i := 0; i < len(h.Sites); i++ {
		if h.Sites[i].DomainName.MatchString(reqSite) {
			s = h.Sites[i]
		}
	}

	// Filter request by require path
	var p *Path
	for i := 0; i < len(s.Paths); i++ {
		if s.Paths[i].Path.MatchString(r.URL.Path) {
			p = s.Paths[i]
		}
	}

	atomic.AddUint64(&p.connectionsNumber, 1)
	fmt.Println(p.OverFlow)
	if p.MaxConnections != 0 && atomic.LoadInt64(&p.currentconnectionsNumber) >= p.MaxConnections {
		switch p.OverFlow {
		case "wait":

			for {
				if atomic.LoadInt64(&p.currentconnectionsNumber) <= p.MaxConnections {
					break
				} else {
					time.Sleep(2 * time.Second)
				}
			}

		case "reject", "":
			w.WriteHeader(500)
			atomic.AddUint64(&p.rejected, 1)
			return
		}

	}

	//	Check is accepted client address
	if p.Accept != nil && !p.Accept.Contains(r.RemoteAddr) {
		w.WriteHeader(500)
		atomic.AddUint64(&p.rejected, 1)
		return
	} else if p.Deny != nil && p.Deny.Contains(r.RemoteAddr) {
		w.WriteHeader(500)
		atomic.AddUint64(&p.rejected, 1)
		return
	}
	atomic.AddInt64(&p.currentconnectionsNumber, 1)
	srvpool := p.Servers
	for i := 0; i < len(p.IPFilter); i++ {
		pool := p.IPFilter[i].Contains(r.RemoteAddr)
		if pool != nil {
			srvpool = pool
			break
		}
	}

	var srv *Server
	var resp *http.Response
	for i := 0; i < len(srvpool.Servers) && resp == nil; i++ {
		srv, err = srvpool.FindServer(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println(err)
			return
		}
		resp, err = srv.Do(strconv.Itoa(p.Toport), r)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if resp == nil {
		return
	}
	fmt.Println(time.Since(start))
	/*re := make([]byte, 0)
	for {
		buff := make([]byte, 500)
		n, err := resp.Body.Read(buff)
		re = append(re, buff[:n]...)
		if err != nil && err.Error() == "EOF" {
			break
		} else if err != nil {
			fmt.Println(err)
			return
		}
	}

	w.Write(re)
	if err != nil {
		fmt.Println(err)
	}*/
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	atomic.AddInt64(&p.currentconnectionsNumber, -1)
	fmt.Println(time.Since(start))
}
