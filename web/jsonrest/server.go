package jsonrest

import (
	"net/http"
	"github.com/geaviation/goboot/logging"
	"github.com/geaviation/goboot/web"
	"github.com/ant0ine/go-json-rest/rest"
)

type JsonRestServer struct {
	web.BasicServer

	Api    *rest.Api
	Router rest.App
}

var log = logging.ContextLogger

func (r *JsonRestServer) Serve(ctx *web.AppContext) {
	r.Ctx = ctx

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

	log.Fatal(http.ListenAndServe(":" + port, r.Api.MakeHandler()))
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
		Timestamp int64 `json:"timestamp"`
	}
	n := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "name")
	v := r.Ctx.Env.GetStringEnv("VCAP_APPLICATION", "version")
	b := r.Ctx.Env.GetStringEnv("build")
	t := web.CurrentTimestamp()
	m := &message{Server: "json-rest", Name: n, Version: v, Build: b, Timestamp: t}

	//r.Handle(m, res, req)
	res.(rest.ResponseWriter).WriteJson(m)
}

func NewJsonRestServer(router ...rest.App) web.Server {
	if len(router) == 0 {
		return &JsonRestServer{Api: rest.NewApi(), Router: nil}
	} else {
		return &JsonRestServer{Api: rest.NewApi(), Router: router[0]}
	}
}
