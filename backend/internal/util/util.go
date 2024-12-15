package util

import (
	"fmt"
	"net/http"

	"github.com/common-nighthawk/go-figure"
)

func ListenHelloFrom(serviceName string) {
	message := figure.NewFigure(serviceName, "", true)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s\n", message)
	})

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
