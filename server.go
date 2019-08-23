package main

import (
  "fmt"
  "net"
  "bufio"
  "strings"
  "time"
  "log"
  "sync"
)

type Server struct {
  Listener net.Listener
  Logger *log.Logger
  Connections *Connections
}

type Connections struct {
  m map[string]*User
  sync.RWMutex
}

type User struct {
  username string
  connection net.Conn
}

func (s *Server) Start() {
  // Accept connections
  for {
    conn, err := s.Listener.Accept()
    if err != nil {
      fmt.Println(err)
      continue
    }

    go s.handleConnection(conn)
  }
}

func (s *Server) listenForMessages(conn net.Conn) {
  for {
    text, err := s.readFromClient(conn)
    if err != nil {
      break
    }
    address := conn.RemoteAddr().String()
    username := s.getUsername(address)

    // Continue if text is empty
    if text == "" {
      continue
    }

    // Client request exit
    // - Remove user from connections map
    // - Close connection
    if text == "-exit" {
      fmt.Printf("Removing %s\n", address)

      s.removeConnection(address)
      break
    }

    // Client request username change
    if text == "-cu" {
      s.handleUsername(conn)
      continue
    }

    // Broadcast message
    s.broadcast(username + ": " + text)
  }
  conn.Close()
}

func (s *Server) handleConnection(conn net.Conn) {
  fmt.Println("Serving " + conn.RemoteAddr().String())

  s.handleUsername(conn)

  go s.listenForMessages(conn)
}

func (s *Server) handleUsername(conn net.Conn) {
  s.writeToClient(conn, "Please enter a username: ")
  username, err := s.readFromClient(conn)
  if err != nil {
    return
  }
  user := &User{}
  user.username = username
  user.connection = conn

  var entering bool
  var oldUsername string

  // Set user to connections map
  address := conn.RemoteAddr().String()
  if s.Connections.m[address] != nil {
    oldUsername = s.Connections.m[address].username
    entering = false
  } else {
    entering = true
  }

  s.addConnection(address, user)

  if entering {
    s.broadcast(user.username + " has entered.")
  } else {
    s.broadcast(oldUsername + " changed username to " + user.username)
  }
}

func (s *Server) writeToClient(conn net.Conn, msg string) {
  _, err := conn.Write([]byte(msg))
  if err != nil {
    fmt.Println("Could not write to " + conn.RemoteAddr().String() + ", removing client")
    // Assuming connection lost, removing client from connections map
    s.removeConnection(conn.RemoteAddr().String())
    return
  }
}

func (s *Server) readFromClient(conn net.Conn) (string, error) {
  netData, err := bufio.NewReader(conn).ReadString('\n')
  if err != nil {
    fmt.Println("Could not read from " + conn.RemoteAddr().String() + ", removing client")
    // Assuming connection lost, removing client from connections map
    s.removeConnection(conn.RemoteAddr().String())
    return "", err
  }

  return strings.TrimSpace(string(netData)), err
}

func (s *Server) broadcast(msg string) {
  currentTime := time.Now()

  // Write to each user connected
  for _, user := range s.Connections.m {
    s.writeToClient(user.connection, currentTime.Format("\n(Mon, Jan 2 2006 - 15:04pm)") + " " + msg + "\n\n")
  }

  s.Logger.Println(msg)
}

func (s *Server) addConnection(address string, user *User) {
  s.Connections.Lock()
  s.Connections.m[address] = user
  s.Connections.Unlock()
}

func (s *Server) removeConnection(address string) {
  fmt.Println("Removing " + address)
  username := s.getUsername(address)

  s.Connections.Lock()
  delete(s.Connections.m, address)
  s.Connections.Unlock()

  s.broadcast(username + " has left.")
}

func (s *Server) getUsername(address string) string {
  s.Connections.RLock()
  username := s.Connections.m[address].username
  s.Connections.RUnlock()

  return username
}
