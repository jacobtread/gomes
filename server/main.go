package server

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/jacobtread/gomes/blaze"
	"log"
	"net"
)

func StartMain() {
	log.Println("GoMES Main Server Starting")

	x509KeyPair, err := tls.X509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Fatalln("Failed to acquire x509KeyPair", err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{x509KeyPair}}

	// Listen using tcp on all addresses with the game port
	t, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", GamePort), config)
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
			log.Println("Failed to accept main connection", err)
			continue
		}
		log.Println("Accepted redirect connection", c)
		go handleConnectionMain(c)
	}
}

func handleConnectionMain(conn net.Conn) {
	buf := blaze.PacketBuff{Buffer: &bytes.Buffer{}}
	bc := blaze.Connection{Conn: conn, PacketBuff: &buf}

	for {
		_, _ = buf.ReadFrom(bc.Conn)
		packet := buf.ReadAllPackets().Front().Value.(blaze.Packet)
		fmt.Println(packet.ToDescriptor())

	}
}
