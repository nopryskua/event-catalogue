package util

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
)

type TypeNamer interface {
	TypeName() string
}

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

func TypeName[T any]() string {
	var v T
	t := reflect.TypeOf(v)

	method, ok := t.MethodByName("TypeName")
	if ok && method.Type.NumIn() == 1 && method.Type.NumOut() == 1 && method.Type.Out(0).Kind() == reflect.String {
		zeroVal := reflect.Zero(t).Interface()
		if typeNamer, ok := zeroVal.(TypeNamer); ok {
			return typeNamer.TypeName()
		}
	}

	return strings.Join([]string{t.PkgPath(), t.Name()}, ".")
}

func Loop(f func() (retry bool), cooldown time.Duration) {
	for {
		if f() {
			time.Sleep(cooldown)

			continue
		}

		return
	}
}
