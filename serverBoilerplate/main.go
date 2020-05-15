package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lSimul/php2go/std/array"

	"github.com/lSimul/php2go/serverBoilerplate/globals"
)

var server = flag.String("S", "", "Run program as a server.")

func main() {
	flag.Parse()

	globals.GET = array.NewString()

	if *server != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/", mainServer)

		// Validate that server address
		log.Fatal(http.ListenAndServe(*server, mux))
	} else {
		mainLCI()
	}
}

func mainServer(w http.ResponseWriter, r *http.Request) {
	globals.W = w
	if r.URL.Path == "/" || r.URL.Path == "/index.php" {
		for k, v := range r.URL.Query() {
			globals.GET.Edit(array.NewScalar(k), v[len(v)-1])
		}
		mainFunc()
		return
	}
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}

func mainLCI() {
	globals.W = os.Stdout
	mainFunc()
}

func mainFunc() {
	fmt.Fprintf(globals.W, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
</head>
<body>
`)
}
