package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/lSimul/php2go/std/array"
)

var server = flag.String("S", "", "Run program as a server.")

type global struct {
	_GET array.String
	W    io.Writer
}

func main() {
	flag.Parse()

	if *server != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/index.php", func(w http.ResponseWriter, r *http.Request) {
			g := &global{_GET: array.NewString()}
			g.W = w
			for k, v := range r.URL.Query() {
				g._GET.Edit(array.NewScalar(k), v[len(v)-1])
			}
			g.mainFunc()
		})

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			g := &global{_GET: array.NewString()}
			g.W = w
			if r.URL.Path != "/" {
				http.FileServer(http.Dir(".")).ServeHTTP(w, r)
			}
			for k, v := range r.URL.Query() {
				g._GET.Edit(array.NewScalar(k), v[len(v)-1])
			}
			g.mainFunc()
		})

		// Validate that server address
		log.Fatal(http.ListenAndServe(*server, mux))
	} else {
		mainLCI()
	}
}

func mainLCI() {
	g := &global{
		W:    os.Stdout,
		_GET: array.NewString(),
	}
	g.mainFunc()
}

func (g *global) mainFunc() {
	fmt.Fprintf(g.W, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
</head>
<body>
`)
}
