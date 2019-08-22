package main

import (
  "fmt"
  "testing"
  "net"
  "log"
  "io/ioutil"
)

func TestUsernameChange(t *testing.T) {
  username := "Ray"
  changeTo := "Kevin"

  go func() {
      conn, err := net.Dial("tcp", ":9000")
      if err != nil {
          t.Fatal(err)
      }

      if _, err := fmt.Fprintf(conn, changeTo + "\n"); err != nil {
          t.Fatal(err)
      }
  }()

  listener, err := net.Listen("tcp", ":9000")
  if err != nil {
      t.Fatal(err)
  }
  defer listener.Close()

  logger := log.New(ioutil.Discard, "", log.LstdFlags)

  conn, err := listener.Accept()
  if err != nil {
    fmt.Println(err)
    t.Fatalf("Could not accept connection.")
  }
  defer conn.Close()

  server := NewServer(listener, logger)

  if len(server.Connections.m) != 0 {
    t.Fatalf("connections should be zero.")
  }

  address := conn.RemoteAddr().String()

  user := &User{}
  user.username = username
  user.connection = conn
  server.addConnection(address, user)

  server.handleUsername(conn)

  if(server.Connections.m[address].username != changeTo) {
     t.Fatalf("Username does not match.")
  }
}

func TestAddConnection(t *testing.T) {
  address := "127.0.0.1:65387"
  username := "Ray"

  listener, err := net.Listen("tcp", ":9000")
  if err != nil {
      t.Fatal(err)
  }
  defer listener.Close()

  logger := log.New(ioutil.Discard, "", log.LstdFlags)

  server := NewServer(listener, logger)

  if len(server.Connections.m) != 0 {
    t.Fatalf("Connections should be zero.")
  }
  user := &User{}
  user.username = username
  user.connection = nil
  server.addConnection(address, user)

  if len(server.Connections.m) != 1 {
    t.Fatalf("Connections should be one.")
  }

  if server.Connections.m[address] == nil {
    t.Fatalf("User was not added to connections map.")
  }
}
