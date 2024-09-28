package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type TCPServer struct {
	ListenAddr string
	Listen     net.Listener
	quitch     chan struct{}
}

func NewTCPServer(laddr string) *TCPServer {
	return &TCPServer{
		ListenAddr: laddr,
		quitch:     make(chan struct{}),
	}
}

func (t *TCPServer) Start() error {
	listener, err := net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	t.Listen = listener
	go t.AcceptLoop()

	<-t.quitch

	close(t.quitch)
	return nil
}

func (t *TCPServer) AcceptLoop() {
	for {
		conn, err := t.Listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("Recieved connection from ", conn.RemoteAddr())
		go t.ReadLoop(conn)
	}
}

// func (t *TCPServer) ReadLoop(conn net.Conn) {
// 	//buf := new(bytes.Buffer)
// 	defer conn.Close()
// 	buf := make([]byte, 2048)
// 	for {
// 		n, err := conn.Read(buf)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Println(err)
// 			continue
// 		}
// 		fmt.Println(string(buf[:n]))
// 		fmt.Printf("Written %d bytes from %v\n", n, conn.RemoteAddr())
// 	}
// }

func (t *TCPServer) ReadLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	//buf := make([]byte, 2048)
	for {
		var size int64
		binary.Read(conn, binary.LittleEndian, &size)
		n, err := io.CopyN(buf, conn, size)
		//n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(buf.Bytes())
		fmt.Printf("Written %d bytes \n", n)
	}
}

func sendFile(size int) error {
	buf := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return err
	}
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}
	binary.Write(conn, binary.LittleEndian, int64(size))
	//n, err := conn.Write(buf)
	_, err = io.CopyN(conn, bytes.NewReader(buf), int64(size))
	if err != nil {
		return err
	}

	fmt.Printf("written %d bytes over the network\n", size)
	return nil
}

func main() {
	go func() {
		time.Sleep(3 * time.Second)
		sendFile(3000)
	}()
	server := NewTCPServer(":3000")
	log.Fatal(server.Start())
}
