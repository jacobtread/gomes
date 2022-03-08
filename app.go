package main

import (
	_ "embed"
	"github.com/jacobtread/gomes/server"
)

func main() {
	go server.StartMain()
	go server.StartRedirector()
}
