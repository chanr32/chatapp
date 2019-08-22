package main

import (
  "fmt"
  "testing"
  "net"
  "io/ioutil"
  "os"
  "encoding/json"
)

func isJSON(s string) bool {
  var js map[string]interface{}
  return json.Unmarshal([]byte(s), &js) == nil
}

func init() {
  os.Setenv("ENV", "Test")
}

func TestUsernameChange(t *testing.T) {
  username := "Ray"

  go func() {
      conn, err := net.Dial("tcp", ":9000")
      if err != nil {
          t.Fatal(err)
      }
      defer conn.Close()

      if _, err := fmt.Fprintf(conn, username + "\n"); err != nil {
          t.Fatal(err)
      }
  }()

  l, err := net.Listen("tcp", ":9000")
  if err != nil {
      t.Fatal(err)
  }
  defer l.Close()

  for {
      conn, err := l.Accept()
      if err != nil {
          return
      }
      defer conn.Close()

      handleUsername(conn)

      if connections.m[conn.RemoteAddr().String()].username != username {
        t.Fatalf("Username does not match.")
      }
      return
  }
}

func TestReadConfig(t *testing.T) {
  data, err := ioutil.ReadFile("test-data/env.json")
  if err != nil {
    t.Fatal("Could not open file.")
  }
  if !isJSON(string(data)) {
    t.Fatal("File is not JSON format.")
  }
}
