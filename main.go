package main

import (
	"fmt"
	"hack-o-ween-site/packages"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	fmt.Println(packages.Test())

	http.ListenAndServe(":9956", nil)
}
