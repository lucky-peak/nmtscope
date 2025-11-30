package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/dist/*
var embeddedFiles embed.FS

// 将 `web/dist` 映射为 FS 子树
func DistFS() http.FileSystem {
	dist, err := fs.Sub(embeddedFiles, "web/dist")
	if err != nil {
		panic(err)
	}
	return http.FS(dist)
}

func SpaHandler() http.Handler {
	fsys := DistFS()
	fileServer := http.FileServer(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试打开文件
		f, err := fsys.Open(r.URL.Path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// fallback: index.html
		index, err := fsys.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		stat, _ := index.Stat()

		http.ServeContent(w, r, "index.html", stat.ModTime(), index)
	})
}
