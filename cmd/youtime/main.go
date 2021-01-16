package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
)

var webDir = flag.String("w", "web/", "web assets directory")
var port = flag.Int("p", 8080, "port for HTTP server")

var monitoringTmpl *template.Template

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	monitoringTmpl = template.Must(template.New(path.Join(*webDir, "monitoring.html")).Parse("html"))
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// c := youtime.NewClient(ctx)

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s%s %s",
			r.Method,
			r.Proto,
			r.Host,
			r.URL,
			r.UserAgent(),
		)
		// fetch stats from youtime client and execute template
		stats := struct{}{}
		err := monitoringTmpl.Execute(w, stats)
		if err != nil {
			log.Print(err)
		}
	}

	http.HandleFunc("/monitoring", handlerFunc)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	// TODO: YouTime NTP server
}
