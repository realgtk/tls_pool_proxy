package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
)

var port, pool string

func main() {
	flag.StringVar(&port, "port", "443", "Usage: poolserver -port port -pool ethash.poolbinance.com:443")
	flag.StringVar(&pool, "pool", "ethash.poolbinance.com:443", "Usage: poolserver -port port -pool ethash.poolbinance.com:443")
	flag.Parse()
	server()
}

func server() {
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Unable to load certificate: %s", err)
		return
	}
	log.Println("Load certificate success...")
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", ":"+port, &config)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("TLS Server Listening: %s", port)
	defer listener.Close()
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go forward(client)
	}
}

func forward(miner net.Conn) {
	conn, err := net.Dial("tcp", pool)
	if err != nil {
		log.Printf("Dial failed: %v", err)
		defer conn.Close()
		return
	}
	log.Printf("Forwarding from %v to %v\n", miner.LocalAddr(), conn.RemoteAddr())

	//miner --> tls --> pool
	go func() {
		defer miner.Close()
		defer conn.Close()
		io.Copy(conn, miner)
	}()

	//pool --> tls --> miner
	go func() {
		defer miner.Close()
		defer conn.Close()
		io.Copy(miner, conn)
	}()
}
