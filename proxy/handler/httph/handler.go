package httph

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/srv"
	"github.com/pkg/errors"
)

var HttpHandlersMutex = &sync.RWMutex{}
var HttpHandlers = make(map[int64]*HttpHandler)

type HttpHandler struct {
	*models.HttpHandler
	stats  models.HandlerStats
	m      sync.RWMutex
	sm     sync.RWMutex
	client http.Client
}

func (h *HttpHandler) AddConn() {
	h.sm.Lock()
	h.stats.Connections++
	h.sm.Unlock()
}

func (h *HttpHandler) RemoveConn() {
	h.sm.Lock()
	h.stats.Connections--
	h.sm.Unlock()
}

// Run HttpHandler job gorutine
func (h *HttpHandler) Listen() {
	go h.listen()
}

// If conn must be secure - check cert
// Get requests and delegate they to handle function
func (h *HttpHandler) listen() {

	if h.Secure {
		// load cert
		certs := make([]tls.Certificate, 0)
		HttpHostsMutex.RLock()
		for _, v := range h.Hosts {

			hst, exists := HttpHosts[v]
			if !exists {
				continue
			}
			certs = append(certs, *hst.Certificate)

		}
		HttpHostsMutex.RUnlock()
		for i := 0; i < len(h.Hosts); i++ {

		}

		TLSConf := &tls.Config{
			Certificates: certs,
		}
		// create server with TLS cert
		server := &http.Server{
			Addr:      fmt.Sprintf(":%d", h.Port),
			TLSConfig: TLSConf,
			Handler:   h,
		}
		config.Glogger.Info("start listen secure", fmt.Sprintf(":%d", h.Port))
		server.ListenAndServeTLS("", "")
	} else {
		// create serfver without cert
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", h.Port),
			Handler: h,
		}
		config.Glogger.Info("start listen ", fmt.Sprintf(":%d", h.Port))
		server.ListenAndServe()
	}
}

// HttpHandler for http.Server
func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// start := time.Now()

	funcName := "ServerHttp()"

	// Filter requset by site domain name
	reqSite, _, err := net.SplitHostPort(r.URL.Host)
	if err == nil {
		fmt.Println(err)
	}
	h.m.RLock()
	var host *HttpHost
	HttpHostsMutex.RLock()
	for _, v := range h.Hosts {
		hst, exists := HttpHosts[v]
		if !exists {
			continue
		}
		hst.m.RLock()
		if hst.HostReg.MatchString(reqSite) {
			host = hst
			hst.m.RUnlock()
			break
		}
		hst.m.RUnlock()
	}
	HttpHostsMutex.RUnlock()
	h.m.RUnlock()
	if host == nil {
		err = errors.New("no such hosts")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	// Filter request by require path

	host.m.RLock()
	var pth *HttpPath
	HttpPathsMutex.RLock()
	for _, v := range host.Paths {
		pt, exists := HttpPaths[v]
		if !exists {
			continue
		}
		pt.m.RLock()
		if pt.PathReg.MatchString(reqSite) {
			pth = pt
			pt.m.RUnlock()
			break
		}
		pt.m.RUnlock()
	}
	HttpPathsMutex.RUnlock()
	host.m.RUnlock()
	if host == nil {
		err = errors.New("no such paths")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	pth.m.RLock()
	defer pth.m.RUnlock()

	//	Check is accepted client address
	// if accept array not empty accepted only addresses contained in this array
	// if accept array empty but deny array not empty denied addresses contained in this array
	// else all accepted
	clinetIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		err = errors.WithMessage(err, "net.SplitHostPort()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	pind := int64(0)
	for _, find := range pth.ProxyFilters {
		srv.ProxyFiltersMutex.RLock()
		f, exists := srv.ProxyFilters[find]
		srv.ProxyFiltersMutex.RUnlock()
		if !exists {
			continue
		}
		pind, err = f.GetPool(clinetIp)
		if err == nil {
			break
		}
	}
	if err != nil {
		err = errors.New("access denied, no filters")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	srv.ServerPoolsMutex.RLock()
	p, exists := srv.ServerPools[pind]
	srv.ServerPoolsMutex.RUnlock()
	if !exists {
		err = errors.New("access denied, no pools")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	sind := p.GetSrv(clinetIp)

	srv.ServersMutex.RLock()
	s, exists := srv.Servers[sind]
	srv.ServersMutex.RUnlock()
	if !exists {
		err = errors.New("access denied, no servers")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	connSocket := fmt.Sprintf("http://%s:%d", s.Address, s.Port)
	// conn, exists := sessions[connSocket]
	// if !exists {
	// 	conn, err = net.Dial("tcp", connSocket)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	sessions[connSocket] = conn
	// }

	req, err := http.NewRequest(r.Method, connSocket, r.Body)
	if err != nil {
		err = errors.WithMessage(err, "http.NewRequest()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	h.m.Lock()
	resp, err := h.client.Do(req)
	if err != nil {
		err = errors.WithMessage(err, "client.Do()")
		config.Glogger.WithField("func", funcName).Error(err)
		h.m.Unlock()
		return
	}
	// http.Transport.Dial
	// req.
	// data, err := httputil.DumpRequest(req, true)
	h.m.Unlock()
	// conn.Write()
	s.AddConn()
	h.AddConn()
	defer func() {
		s.RemoveConn()
		h.RemoveConn()
	}()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		err = errors.WithMessage(err, "io.Copy()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	// fmt.Println(time.Since(start))
}
