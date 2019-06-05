package jsonrest

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gostones/goboot/logging"
	"github.com/gostones/goboot/web"
	"net/http"
)

type JsonRestServer struct {
	web.BasicServer

	Api    *rest.Api
	Router rest.App
}

var log = logging.Logger()

func (r *JsonRestServer) Serve() {
	port := r.Port()

	//
	r.Api.Use(rest.DefaultDevStack...)

	//
	if r.Router == nil {
		var err error
		r.Router, err = rest.MakeRouter(
			rest.Get("/", HandlerAdapter(r.home)),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	r.Api.SetApp(r.Router)

	log.Infof("Server listening on port: %s", port)

	log.Fatal(http.ListenAndServe(":"+port, r.Api.MakeHandler()))
}

func HandlerAdapter(handler func(http.ResponseWriter, *http.Request)) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		handler(res.(http.ResponseWriter), req.Request)
	}
}

func (r *JsonRestServer) home(res http.ResponseWriter, req *http.Request) {
	type message struct {
		Server    string `json:"server"`
		Name      string `json:"name"`
		Version   string `json:"version"`
		Build     string `json:"build"`
		Timestamp int64  `json:"timestamp"`
	}
	n := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "name")
	v := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "version")
	b := r.Ctx.Env.GetStringEnv("build")
	t := web.CurrentTimestamp()
	m := &message{Server: "json-rest", Name: n, Version: v, Build: b, Timestamp: t}

	//r.Handle(m, res, req)
	res.(rest.ResponseWriter).WriteJson(m)
}

func NewJsonRestServer(router ...rest.App) *JsonRestServer {
	ctx := web.CreateAppContext()

	if len(router) == 0 {
		return &JsonRestServer{BasicServer: web.BasicServer{Ctx: ctx}, Api: rest.NewApi(), Router: nil}
	} else {
		return &JsonRestServer{BasicServer: web.BasicServer{Ctx: ctx}, Api: rest.NewApi(), Router: router[0]}
	}
}
