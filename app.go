package main

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"fmt"
	"github.com/jacobtread/gomes/blaze"
	"log"
	"net"
)

//go:embed cert/cert.pem
var certFile []byte

//go:embed cert/key.pem
var keyFile []byte

func main() {

	log.Println("GoMES Starting")

	const port = 14219 // The mass effect 3 game server port

	x509KeyPair, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalln("Failed to acquire x509KeyPair", err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{x509KeyPair}}

	// Listen using tcp on all addresses with the game port
	t, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), config)
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
			log.Println("Failed to accept connection", err)
			continue
		}
		conn := blaze.Connection{Conn: c}
		log.Println("Accepted connection", conn)
		go HandleConnection(&conn)
	}
}

func HandleConnection(conn *blaze.Connection) {
	buf := blaze.PacketBuff{Buffer: &bytes.Buffer{}}
	conn.PacketBuff = buf

	for {
		_, _ = buf.ReadFrom(conn)

		p := buf.ReadPacket()
		log.Println(p)
	}
}
