package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var clients = make(map[string]net.Conn)
var usernames = make(map[string]string)
var leaving = make(chan message)
var messages = make(chan message)

type message struct {
	text    string
	address string
}

func handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()

	clients[addr] = conn

	input := bufio.NewScanner(conn)
	firstSent := true

	for input.Scan() {
		if firstSent {
			usernames[addr] = input.Text()
			fmt.Println(usernames[addr] + " (" + addr + ") joined.")
			messages <- newMessage(" joined.", conn)
			firstSent = false
		} else {
			messages <- newMessage(": "+input.Text(), conn)
		}
	}

	delete(clients, addr)
	leaving <- newMessage(" has left.", conn)
	fmt.Println(usernames[addr] + " (" + addr + ") has left.")

	conn.Close()
}

func newMessage(msg string, conn net.Conn) message {
	addr := conn.RemoteAddr().String()
	user := usernames[addr]

	return message{
		text:    user + msg,
		address: addr,
	}
}

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			for _, conn := range clients {
				if msg.address == conn.RemoteAddr().String() {
					continue
				}

				fmt.Fprintln(conn, msg.text)
			}

		case msg := <-leaving:
			for _, conn := range clients {
				fmt.Fprintln(conn, msg.text)
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		return
	}

	go broadcaster()
	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Print(err)
			continue
		}

		go handleConnection(conn)
	}
}
