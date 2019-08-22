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

type Server struct {
  Listener net.Listener
  Logger *log.Logger
  Connections *Connections
}

type Config struct {
  Ip, Port, LogFile string
}

type Connections struct {
  m map[string]*User
  sync.RWMutex
}

type User struct {
  username string
  connection net.Conn
}

func NewServer(listener net.Listener, logger *log.Logger) *Server {
  return &Server{
    Listener: listener,
    Logger: logger,
    Connections: &Connections{m: make(map[string]*User)},
  }
}

func setDefault(config *Config) {
  config.Ip = "localhost"
  config.Port = "9000"
  config.LogFile = "chat.log"
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

func main() {
  config := &Config{}

  // Read config from file
  configFile, err := os.Open("env.json")
  defer configFile.Close()
  if err != nil {
    fmt.Println("Warn: config file does not exist, falling back to default")
    setDefault(config)
  } else {
    decoder := json.NewDecoder(configFile)
    err := decoder.Decode(config)
    if err != nil {
      fmt.Println("Warn: failed decoding json file, falling back to default")
      setDefault(config)
    }
  }

  // Open/Create Log file
  logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
  	fmt.Println(err)
  }
  defer logFile.Close()

  logger := log.New(logFile, "chat-app: ", log.LstdFlags)

  // Create listener
  listener, err := net.Listen("tcp", config.Ip + ":" + config.Port)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer listener.Close()

  // Start server
  server := NewServer(listener, logger)
  server.Start()
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

  // Set user to connections map
  address := conn.RemoteAddr().String()
  s.addConnection(address, user)
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

  s.broadcast(user.username + " has entered.")
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
