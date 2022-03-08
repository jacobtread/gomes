package server

import (
	"bytes"
	"fmt"
	"github.com/jacobtread/gomes/blaze"
	"io/ioutil"
	"log"
	"net"
)

func StartRedirector() {
	log.Println("GoMES Redirector Starting")

	// Listen using tcp on all addresses with the game port
	t, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", RedirectorPort))
	if err != nil {
		panic(err)
		return
	}
	// Deferred closing of the listener
	defer func(t net.Listener) { _ = t.Close() }(t)

	log.Println("Waiting for connections...")

	for {
		c, err := t.Accept()
		if err != nil {
			log.Println("Failed to accept redirector connection", err)
			continue
		}
		log.Println("Accepted redirect connection")
		go handleConnectionRedirect(c)
	}
}

func handleConnectionRedirect(conn net.Conn) {
	buf := blaze.PacketBuff{}
	bc := blaze.Connection{Conn: conn, PacketBuff: &buf}
	for {
		b, err := ioutil.ReadAll(conn)
		if err != nil {
			panic(err)
		}
		bc.Buffer = bytes.NewBuffer(b)
		if buf.Len() > 0 {
			packet := buf.ReadPacket()
			fmt.Println(packet.ToDescriptor())

			return
		}
	}
}
