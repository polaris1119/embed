package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/polaris1119/embed"
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", hashfs.FileServer(embed.Fsys)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tpl, err := template.New("index.html").ParseFiles("template/index.html")
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		err = tpl.Execute(w, map[string]interface{}{
			"mainjs": embed.Fsys.HashName("static/main.js"),
		})
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
