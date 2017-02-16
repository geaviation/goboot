package web

import (
	"net/http"
	"github.com/geaviation/goboot/config"
	"github.com/geaviation/goboot/logging"
	"encoding/json"
	"fmt"
	"time"
	"github.com/tylerb/graceful"
)

type Server interface {
	Serve(ctx *AppContext)
}

type AppContext struct {
	Env *config.Settings
}

type BasicServer struct {
	Ctx    *AppContext

	Router *http.ServeMux
}

var log = logging.ContextLogger

func createAppContext() *AppContext {
	p := config.NewSettings()

	ctx := AppContext{
		Env: p,
	}
	return &ctx
}

func CurrentTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (r *BasicServer) Port() string {
	port := r.Ctx.Env.GetStringEnv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func (r *BasicServer) Serve(ctx *AppContext) {
	r.Ctx = ctx

	if r.Router == nil {
		r.Router = http.NewServeMux()
		r.Router.HandleFunc("/", r.home)
	}

	r.Start(r.Router)
}

func (r *BasicServer) Start(handler *http.ServeMux) {
	port := r.Port()

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}

	log.Infof("Server listening on port: %s", port)

	log.Fatal(server.ListenAndServe())
}

func (r *BasicServer) home(res http.ResponseWriter, req *http.Request) {
	type message struct {
		Server    string `json:"server"`
		Name      string `json:"name"`
		Version   string `json:"version"`
		Build     string `json:"build"`
		Timestamp int64 `json:"timestamp"`
	}
	n := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "name")
	v := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "version")
	b := r.Ctx.Env.GetStringEnv("build")
	t := CurrentTimestamp()
	m := &message{Server: "basic", Name: n, Version: v, Build: b, Timestamp: t}

	r.Handle(m, res, req)
}

func (r *BasicServer) Handle(m interface{}, res http.ResponseWriter, req *http.Request) {
	Handle(m, res, req)
}

func Handle(m interface{}, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Errorf("Handle: %s", r)
		}
	}()

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	b, _ := json.Marshal(m)
	fmt.Fprintf(res, string(b))
}

func NewBasicServer(router ...*http.ServeMux) Server {
	if len(router) == 0 {
		return &BasicServer{Router: nil}
	} else {
		return &BasicServer{Router: router[0]}
	}
}

func Run(s ...Server) {
	ctx := createAppContext()

	if len(s) == 0 {
		bs := NewBasicServer()
		bs.Serve(ctx)
	} else {
		s[0].Serve(ctx)
	}

	log.Error("Server exiting.")
}