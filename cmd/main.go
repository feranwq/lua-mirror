package main

import (
	"compress/flate"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/feranwq/lua-mirror/webrouter"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version        = "0.0.1"
	listenAddress  = kingpin.Flag("web.listen-address", "The address to listen on for web interface.").Default(":8080").String()
	luarockServer  = kingpin.Flag("luarock.server", "Luarock server address").Default("http://luafr.org/luarocks").String()
	requestTimeout = kingpin.Flag("notify.timeout", "Timeout for luarock server").Default("5s").Duration()
	dataDir        = kingpin.Flag("data.dir", "Data directory").Default(".").String()
)

func main() {
	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	level.Info(logger).Log("msg", "Starting luarock-mirror server", "version", version)
	level.Info(logger).Log("msg", "Luarock server address", "luarock.server", *luarockServer)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	compressor := middleware.NewCompressor(flate.DefaultCompression)
	r.Use(compressor.Handler)

	workDir, _ := os.Getwd()
	dataDir := filepath.Join(workDir, *dataDir)
	downloadQueue := make(map[string]int, 10)

	luarockMirror := &webrouter.LuaMirror{
		Logger:        logger,
		Root:          dataDir,
		Path:          "/",
		LuarockServer: *luarockServer,
		DownlowdQueue: downloadQueue,
	}

	r.Mount("/", luarockMirror.Routes())

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, r); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}

}

