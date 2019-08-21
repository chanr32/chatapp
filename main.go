package main

import (
  "fmt"
  "os"
  "net"
  "bufio"
  "strings"
  "time"
  "log"
  "encoding/json"
)

var connections = make(map[string]User)
var serverVariables ServerVariable

type ServerVariable struct {
  Port, Ip, LogFile string
  Logger *log.Logger
}

type User struct {
  username string
  connection net.Conn
}

func main() {
  file, _ := os.Open("env.json")
  defer file.Close()

  decoder := json.NewDecoder(file)
  err := decoder.Decode(&serverVariables)
  if err != nil {
    // set to default
    serverVariables.Ip = "localhost"
    serverVariables.Port = "9000"
    serverVariables.LogFile = "chat.log"
  }

  ln, err := net.Listen("tcp", serverVariables.Ip + ":" + serverVariables.Port)
  if err != nil {
    panic(err)
  }
  defer ln.Close()

  f, err := os.OpenFile(serverVariables.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
  	log.Println(err)
  }
  defer f.Close()

  serverVariables.Logger = log.New(f, "chat-app: ", log.LstdFlags)

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

  serverVariables.Logger.Println(msg)
}
