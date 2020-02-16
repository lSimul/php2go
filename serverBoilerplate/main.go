package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var W io.Writer

var server = flag.String("S", "", "Run program as a server.")

func main() {
	flag.Parse()

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
