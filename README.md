# DM883 - Project
Decentralized chat application

# Run the apps
## Relay
1. Open a new terminal
2. Navigate to the relay folder
3. Compile the app
```
go build -o myrelay
```
4. Run the app
```
./myrelay
```

## Client
1. Open a new terminal
2. Navigate to the client folder
3. Compile the app
```
go build -o myclient
```
4. Run the app with an available port
```
./myclient 8080
```
5. In a new terminal, run a client on another port
```
./myclient 8081
```
6. In both clients, add the relay by specifying IP 127.0.0.1 and whatever port was printed when the relay started.
7. Create a group in client 1
8. Search for the group name in client 2
9. Both clients can join the group and start chatting.*
10. Unfortunately you need to manually update the window or send a message to refresh new messages from the other client.