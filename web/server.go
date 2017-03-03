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
	Serve()
}

type AppContext struct {
	Env *config.Settings
}

type BasicServer struct {
	Ctx    *AppContext

	Router *http.ServeMux
}

var ContentType = struct {
	JSON string
	HTML string
	JS   string
	CSS  string
	BIN  string
}{
	JSON: "application/json",
	HTML: "text/html",
	JS:   "application/javascript",
	CSS:  "text/css",
	BIN:  "application/octet-stream",
}

var log = logging.Logger()

func CreateAppContext() *AppContext {
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

func (r *BasicServer) Serve() {
	if r.Router == nil {
		r.Router = http.NewServeMux()
		r.Router.HandleFunc("/", r.home)
	}

	r.Start()
}

func (r *BasicServer) Start() {
	port := r.Port()

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    ":" + port,
			Handler: r.Router,
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

	r.HandleJson(m, res, req)
}

func (r *BasicServer) HandleJson(m interface{}, res http.ResponseWriter, req *http.Request) {
	HandleJson(m, res, req)
}

func HandleJson(m interface{}, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Errorf("Handle: %s", r)
		}
	}()

	res.Header().Set("Content-Type", ContentType.JSON)
	res.WriteHeader(http.StatusOK)

	b, _ := json.Marshal(m)
	fmt.Fprintf(res, string(b))
}

func NewBasicServer(router ...*http.ServeMux) *BasicServer {
	ctx := CreateAppContext()

	if len(router) == 0 {
		return &BasicServer{Ctx: ctx, Router: nil}
	} else {
		return &BasicServer{Ctx: ctx, Router: router[0]}
	}
}

func Run(s ...Server) {
	if len(s) == 0 {
		bs := NewBasicServer()
		bs.Serve()
	} else {
		s[0].Serve()
	}

	log.Error("Server exiting.")
}