package gorilla

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"github.com/geaviation/goboot/logging"
	"github.com/geaviation/goboot/web"
)

type GorillaServer struct {
	ctx *web.AppContext
}

var log = logging.ContextLogger

func currentTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (r *GorillaServer) port() string {
	port := r.ctx.Env.GetStringEnv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func (r *GorillaServer) Serve(ctx *web.AppContext) {
	r.ctx = ctx

	port := r.port()

	mux := mux.NewRouter()

	mux.HandleFunc("/", r.home)

	log.Infof("Server listening on port: %s", port)

	http.ListenAndServe(":" + port, mux)
}

func (r *GorillaServer) home(res http.ResponseWriter, req *http.Request) {
	type message struct {
		Server      string `json:"server"`
		Name      string `json:"name"`
		Version   string `json:"version"`
		Build     string `json:"build"`
		Timestamp int64 `json:"timestamp"`
	}
	n := r.ctx.Env.GetStringEnv("VCAP_APPLICATION", "name")
	v := r.ctx.Env.GetStringEnv("VCAP_APPLICATION", "version")
	b := r.ctx.Env.GetStringEnv("build")
	t := currentTimestamp()
	m := &message{Server: "gorilla", Name: n, Version: v, Build: b, Timestamp: t}

	r.Handle(m, res, req)
}

func (r *GorillaServer) Handle(m interface{}, res http.ResponseWriter, req *http.Request) {
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

func NewGorillaServer() web.Server {
	return &GorillaServer{}
}