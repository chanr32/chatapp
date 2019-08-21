## go chat app

#### Extra Features
- Read from config file
- Client can change username
- Log messages to file

#### How to use
```
go run main.go
```

Clients dialing in will first be prompted to enter their desired username. Clients are able to change their username by sending command **-cu**. Disconnect by sending **-exit**.

Server reads from **env.json** file for configurations. If no file is provided default settings will be used. Defaults are as follow:

```json
{
  "Port": "9000",
  "Ip": "localhost",
  "LogFile": "chat.log"
}
```

Server logs all clients entering and leaving as well as their messages. Location of log file can be set in **env.json**.
