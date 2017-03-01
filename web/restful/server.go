package restful

import (
	"net/http"
	"github.com/geaviation/goboot/logging"
	"github.com/geaviation/goboot/web"
	"github.com/emicklei/go-restful"
)

type RestfulServer struct {
	web.BasicServer

	Router *restful.WebService
}

var log = logging.ContextLogger

func (r *RestfulServer) Serve() {
	port := r.Port()

	//
	if r.Router == nil {
		r.Router = new(restful.WebService)
		r.Router.Route(r.Router.GET("/").To(HandlerAdapter(r.home)))
	}

	restful.Add(r.Router)

	log.Infof("Server listening on port: %s", port)

	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func HandlerAdapter(handler func(http.ResponseWriter, *http.Request)) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler(res, req.Request)
	}
}

func (r *RestfulServer) home(res http.ResponseWriter, req *http.Request) {
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
	m := &message{Server: "restful", Name: n, Version: v, Build: b, Timestamp: t}

	r.Handle(m, res, req)
}

func NewRestfulServer(router ...*restful.WebService) *RestfulServer {
	ctx := web.CreateAppContext()

	if len(router) == 0 {
		return &RestfulServer{BasicServer: web.BasicServer{Ctx: ctx}, Router: nil}
	} else {
		return &RestfulServer{BasicServer: web.BasicServer{Ctx: ctx}, Router: router[0]}
	}
}

