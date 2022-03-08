package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/jacobtread/gomes/blaze"
	"log"
	"net"
)

func StartRedirector() {
	log.Println("GoMES Redirector Starting")

	x509KeyPair, err := tls.X509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Fatalln("Failed to acquire x509KeyPair", err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{x509KeyPair}}

	// Listen using tcp on all addresses with the game port
	t, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", RedirectorPort), config)
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
		conn := blaze.Connection{Conn: c}
		log.Println("Accepted connection", conn)
		go handleConnectionRedirector(&conn)
	}
}

func handleConnectionRedirector(conn *blaze.Connection) {
	buf := blaze.PacketBuff{Buffer: &bytes.Buffer{}}
	conn.PacketBuff = buf

	for {
		_, _ = buf.ReadFrom(conn)
		packet := buf.ReadAllPackets().Front().Value.(blaze.Packet)
		println(packet)

	}
}
