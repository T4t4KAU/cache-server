package servers

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"

	"cache-server/caches"
	"cache-server/router"
)

type HTTPServer struct {
	cache *caches.Cache
}

// 创建HTTP服务器
func NewHTTPServer(cache *caches.Cache) *HTTPServer {
	return &HTTPServer{
		cache: cache,
	}
}

func (server *HTTPServer) Run(address string) error {
	return http.ListenAndServe(address, server.routerHandler())
}

func wrapUriWithVersion(uri string) string {
	return path.Join("/", APIVersion, uri)
}

func (server *HTTPServer) routerHandler() *router.Router {
	r := router.New()
	r.GET(wrapUriWithVersion("/cache/:key"), server.getHandler)
	r.PUT(wrapUriWithVersion("/cache/:key"), server.setHandler)
	r.DELETE(wrapUriWithVersion("/cache/:key"), server.deleteHandler)
	r.GET(wrapUriWithVersion("/status"), server.statusHandler)
	return r
}

func (server *HTTPServer) setHandler(ctx *router.Context) {
	key := ctx.Params.ByName("key")
	value, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ttl, err := parseTTL(ctx.Req)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = server.cache.SetWithTTL(key, value, ttl)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusRequestEntityTooLarge)
		ctx.Writer.Write([]byte("Error: " + err.Error()))
		return
	}
	ctx.Writer.WriteHeader(http.StatusCreated)
}

func parseTTL(request *http.Request) (int64, error) {
	ttls, ok := request.Header["Ttl"]
	if !ok || len(ttls) < 1 {
		return caches.NeverDie, nil
	}
	return strconv.ParseInt(ttls[0], 10, 64)
}

func (server *HTTPServer) getHandler(ctx *router.Context) {
	key := ctx.Params.ByName("key")
	value, ok := server.cache.Get(key)
	if !ok {
		ctx.Writer.WriteHeader(http.StatusNotFound)
		return
	}
	ctx.Writer.Write(value)
}

func (server *HTTPServer) deleteHandler(ctx *router.Context) {
	key := ctx.Params.ByName("key")
	err := server.cache.Delete(key)
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (server *HTTPServer) statusHandler(ctx *router.Context) {
	status, err := json.Marshal(server.cache.Status())
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx.Writer.Write(status)
}
