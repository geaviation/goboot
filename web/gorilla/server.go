package gorilla

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/geaviation/goboot/logging"
	"github.com/geaviation/goboot/web"
)

type GorillaServer struct {
	web.BasicServer
}

var log = logging.ContextLogger

func (r *GorillaServer) Serve(ctx *web.AppContext) {
	r.Ctx = ctx

	port := r.Port()

	mux := mux.NewRouter()

	mux.HandleFunc("/", r.home)

	log.Infof("Server listening on port: %s", port)

	http.ListenAndServe(":" + port, mux)
}

func (r *GorillaServer) home(res http.ResponseWriter, req *http.Request) {
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
	t := web.CurrentTimestamp()
	m := &message{Server: "gorilla", Name: n, Version: v, Build: b, Timestamp: t}

	r.Handle(m, res, req)
}

func NewGorillaServer() web.Server {
	return &GorillaServer{}
}