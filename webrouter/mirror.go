package webrouter

import (
	"github.com/feranwq/lua-mirror/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LuaMirror struct {
	Logger        log.Logger
	Root          string
	Path          string
	Datadir       string
	LuarockServer string
	DownlowdQueue map[string]int
}

func (lm *LuaMirror) Routes() chi.Router {
	path := lm.Path
	r := chi.NewRouter()
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}

	path += "*"
	r.Get(path, lm.MirrorServer)
	return r

}

func (lm *LuaMirror) MirrorServer(w http.ResponseWriter, r *http.Request) {
	var root http.FileSystem = http.Dir(lm.Root)
	logger := lm.Logger
	filePath := lm.Root + r.URL.Path
	serverPath := lm.LuarockServer + r.URL.Path

	if r.URL.Path != "/" && !strings.HasSuffix(r.URL.Path, ".zip") && lm.CheckFileModified(serverPath) {
		if _, exist := lm.DownlowdQueue[filePath]; !exist {
			lm.DownlowdQueue[filePath] = 0
			go lm.DownloadFromUrl(serverPath)
		}
		level.Info(logger).Log("msg", "File not found or modified, cache it and redirect to mirror server", "Redirect url", serverPath, "downloadQueue", len(lm.DownlowdQueue))
		s := http.RedirectHandler(serverPath, 302)
		s.ServeHTTP(w, r)
	}

	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServer(root))
	fs.ServeHTTP(w, r)
}

func (lm *LuaMirror) DownloadFromUrl(url string) {
	logger := lm.Logger
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	filePath := filepath.Join(lm.Root, fileName)

	defer delete(lm.DownlowdQueue, filePath)
	level.Info(logger).Log("msg", "Download started", "Downloadurl", url, "Downloadto", filePath)

	output, err := os.Create(filePath + ".tmp")
	if err != nil {
		level.Error(logger).Log("msg", "Error while creating", "DownloadFile", fileName, "err", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		level.Error(logger).Log("msg", "Error while creating", "DownloadFile", fileName, "err", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		level.Error(logger).Log("msg", "Error while creating", "DownloadFile", fileName, "err", err)
		return
	}
	level.Info(logger).Log("msg", "Download success", "DownloadFile", fileName, "Byte", n)

	os.Rename(filePath+".tmp", filePath)
	if err != nil {
		level.Error(logger).Log("msg", "Error while rename file", "DownloadFile", fileName, "err", err)
		return
	}

}

func (lm *LuaMirror) CheckFileModified(url string) bool {
	logger := lm.Logger
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	filePath := filepath.Join(lm.Root, fileName)

	if !utils.FileExists(filePath) {
		level.Info(logger).Log("msg", "File not found", "file", fileName)
		return true
	}

	response, err := http.Get(url)
	if err != nil {
		level.Error(logger).Log("msg", "Url check failed", "url", url, "err", err)
		return false
	}
	defer response.Body.Close()

	if t, ok := response.Header["Last-Modified"]; ok {
		destTime, err := time.Parse(time.RFC1123, t[0])
		if err != nil {
			level.Error(logger).Log("msg", "Last-Modified header not found", "file", fileName, "err", err)
			return false
		}

		stat, err := os.Lstat(filePath)
		if err != nil {
			level.Error(logger).Log("msg", "File stat not found", "file", fileName, "err", err)
			return false
		}

		localTime := stat.ModTime()
		if destTime.Unix() > localTime.Unix() {
			level.Info(logger).Log("msg", "File modified", "file", fileName)
			return true
		}
	}

	return false
}
