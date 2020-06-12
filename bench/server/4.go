package main

import (
	"flag"
	"fmt"
	"github.com/lSimul/php2go/std/array"
	"io"
	"log"
	"net/http"
	"os"
)

var server *string = flag.String("S", "", "Run program as a server.")
var file *string = flag.String("f", "", "Run designated file.")

type global struct {
	W    io.Writer
	_GET array.String
}

func main() {
	flag.Parse()

	if *server != "" {
		mux := http.NewServeMux()

		mux.HandleFunc("/4.php", func(w http.ResponseWriter, r *http.Request) {
			g := &global{
				_GET: array.NewString(),
				W:    w,
			}
			for k, v := range r.URL.Query() {
				g._GET.Edit(array.NewScalar(k), v[len(v)-1])
			}
			g.mainFunc0()
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(http.Dir(".")).ServeHTTP(w, r)
		})
		// Validate that server address
		log.Fatal(http.ListenAndServe(*server, mux))
	} else {
		mainCLI()
	}
}

func mainCLI() {
	g := &global{
		_GET: array.NewString(),
		W:    os.Stdout,
	}
	switch *file {

	case "4.php":
		g.mainFunc0()

	default:
		g.mainFunc0()
	}
}
func (g *global) mainFunc0() {
	fmt.Fprintf(g.W, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>3</title>
</head>
<body>
	`)
	if g._GET.Count() > 0 {
		fmt.Fprintf(g.W, "<table>")
		fmt.Fprintf(g.W, "\n")
		fmt.Fprintf(g.W, "<tr><th>Key</th><th>Value</th>")
		fmt.Fprintf(g.W, "\n")
		for _, pair := range g._GET.KeyIter() {
			k := pair.K
			v := pair.V
			fmt.Fprintf(g.W, fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", k, v))
		}
		fmt.Fprintf(g.W, "</table>")
	} else {
		fmt.Fprintf(g.W, "<p>No GET params</p>")
	}
	fmt.Fprintf(g.W, `</body>
</html>
`)
}
