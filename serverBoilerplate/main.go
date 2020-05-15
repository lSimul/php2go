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

var W io.Writer

// Intentionally bad name, same name as the
// PHP version. I will do renaming later on.
// This is just to make translation now easier,
// straight match.
var _GET array.String

var server = flag.String("S", "", "Run program as a server.")

func main() {
	flag.Parse()

	_GET = array.NewString()

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
	W = w
	if r.URL.Path == "/" || r.URL.Path == "/index.php" {
		for k, v := range r.URL.Query() {
			_GET.Edit(array.NewScalar(k), v[len(v)-1])
		}
		mainFunc()
		return
	}
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}

func mainLCI() {
	W = os.Stdout
	mainFunc()
}

func mainFunc() {
	fmt.Fprintf(W, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
</head>
<body>
`)
}
