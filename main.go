package main

import (
  "fmt"
  "os"
  "net"
  "log"
  "encoding/json"
)

type Config struct {
  Ip, Port, LogFile string
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
