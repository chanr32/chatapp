package main

import (
  "fmt"
  "net"
  "bufio"
  "strings"
  "time"
)

var connections = make(map[string]User)

type User struct {
  username string
  connection net.Conn
}

func main() {
  ln, err := net.Listen("tcp", ":9000")
  if err != nil {
    panic(err)
  }
  defer ln.Close()

  for {
    conn, err := ln.Accept()
    if err != nil {
      panic(err)
    }

    go handleConnection(conn)
  }
}

func handleUsername(c net.Conn) {
  c.Write([]byte("Please enter a username: "))
  netData, err := bufio.NewReader(c).ReadString('\n')
  if err != nil {
    fmt.Println(err)
    return
  }

  username := strings.TrimSpace(string(netData))
  var user = new(User)
  user.username = username
  user.connection = c
  if _, ok := connections[c.RemoteAddr().String()]; ok {
    broadcast(connections[c.RemoteAddr().String()].username + " has changed their username to " + username)
  } else {
    broadcast(username + " has entered.")
  }

  connections[c.RemoteAddr().String()] = *user

}

func handleConnection(c net.Conn) {
  fmt.Printf("Serving %s\n", c.RemoteAddr().String())

  handleUsername(c)

  for {
          netData, err := bufio.NewReader(c).ReadString('\n')
          if err != nil {
            fmt.Println(err)
            return
          }

          var address = c.RemoteAddr().String()
          var username = connections[address].username
          text := strings.TrimSpace(string(netData))
          if text == "" {
            continue
          }

          if text == "-exit" {
            broadcast(username + " has left.")
            break
          }

          if text == "-changename" {
            handleUsername(c)
            continue
          }

          broadcast(username + ": " + text)
  }
  c.Close()
}

func broadcast(msg string) {
  currentTime := time.Now()

  for _, user := range connections {
    user.connection.Write([]byte(currentTime.Format("\n(Mon, Jan 2 2006 - 15:04pm)") + " " + msg + "\n\n"))
  }
}
