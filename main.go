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
  "sync"
)

var connections = Connections{m: make(map[string]*User)}
var serverVariables ServerVariable

type Connections struct {
  m map[string]*User
  sync.RWMutex
}

type ServerVariable struct {
  Port, Ip, LogFile string
  Logger *log.Logger
}

type User struct {
  username string
  connection net.Conn
}

func setDefault() {
  serverVariables.Ip = "localhost"
  serverVariables.Port = "9000"
  serverVariables.LogFile = "chat.log"
}

func main() {
  file, err := os.Open("env.json")
  defer file.Close()
  if err != nil {
    fmt.Println("Warn: config file does not exist, falling back to default")
    setDefault()
  } else {
    decoder := json.NewDecoder(file)
    err := decoder.Decode(&serverVariables)
    if err != nil {
      fmt.Println("Warn: failed decoding json file, falling back to default")
      setDefault()
    }
  }

  ln, err := net.Listen("tcp", serverVariables.Ip + ":" + serverVariables.Port)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer ln.Close()

  f, err := os.OpenFile(serverVariables.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
  	fmt.Println(err)
  }
  defer f.Close()

  serverVariables.Logger = log.New(f, "chat-app: ", log.LstdFlags)

  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Println(err)
      continue
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

  connections.RLock()
  if _, ok := connections.m[c.RemoteAddr().String()]; ok {
    broadcast(connections.m[c.RemoteAddr().String()].username + " has changed their username to " + username)
  } else {
    broadcast(username + " has entered.")
  }
  connections.RUnlock()

  connections.Lock()
  connections.m[c.RemoteAddr().String()] = user
  connections.Unlock()
}

func handleConnection(c net.Conn) {
  fmt.Printf("Serving %s\n", c.RemoteAddr().String())

  handleUsername(c)

  for {
    netData, err := bufio.NewReader(c).ReadString('\n')
    if err != nil {
      fmt.Println(err)
      break
    }

    var address = c.RemoteAddr().String()

    connections.RLock()
    var username = connections.m[address].username
    connections.RUnlock()

    text := strings.TrimSpace(string(netData))
    if text == "" {
      continue
    }

    if text == "-exit" {
      fmt.Printf("Removing %s\n", address)

      connections.Lock()
      delete(connections.m, address)
      connections.Unlock()

      broadcast(username + " has left.")
      break
    }

    if text == "-cu" {
      handleUsername(c)
      continue
    }

    broadcast(username + ": " + text)
  }
  c.Close()
}

func broadcast(msg string) {
  currentTime := time.Now()

  for address, user := range connections.m {
    _, err := user.connection.Write([]byte(currentTime.Format("\n(Mon, Jan 2 2006 - 15:04pm)") + " " + msg + "\n\n"))
    if err != nil {
      fmt.Println("Could not write to " + address + ", connection dropped.")
      fmt.Println("Removing " + address)
      connections.Lock()
      delete(connections.m, address)
      connections.Unlock()
    }
  }

  serverVariables.Logger.Println(msg)
}
