package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

var hub Hub

func main() {
	hub = Hub{}
	hub.Add("projector", "front", map[string]string{
		"address": "10.0.0.30:8000",
	})
	hub.Add("projector", "rear", map[string]string{
		"address": "10.0.0.31:8000",
	})
	hub.Add("decoder", "desk", map[string]string{
		"address": "http://10.0.0.40:8080",
	})
	//hub.Add("decoder", "pc", map[string]string{
	//	"address": "http://10.0.0.40:8080",
	//})
	hub.Add("encoder", "cam", map[string]string{
		"address": "10.0.0.20:5961",
	})
	hub.Add("encoder", "pc", map[string]string{
		"address": "10.0.0.21:5961",
	})

	http.HandleFunc("/", serveStatic)
	http.HandleFunc("/ws", serveWebsocket)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}

	// never give up
	select {}
}

//go:embed public/*
var statics embed.FS
func serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}

	t, err := template.ParseFS(statics, "public" + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		fmt.Printf("ERROR template.ParseFS: %v\n", err)
		return
	}

	// Content-Type
	switch true {
	case strings.HasSuffix(r.URL.Path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(r.URL.Path, ".map"):
		w.Header().Set("Content-Type", "application/json")
	case strings.HasSuffix(r.URL.Path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/html")
	}

	// HTTP 200
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Server error", 500)
		fmt.Printf("ERROR tpl.Execute: %v\n", err)
		return
	}

	return
}

func serveWebsocket(w http.ResponseWriter, r *http.Request) {
	client := Client{}
	client.Serve(w, r)
}