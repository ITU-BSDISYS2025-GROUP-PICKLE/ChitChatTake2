# BSDISYS1KU Group pickle || Mandatory activity 3: ChitChat (Take 2)
## How to run the program
1. In four consoles, start by CD'ing to the ChitChat directory
2. In one console, run "go run server/server.go"
3. In the other three consoles, run "go run client/client.go"
6. You can now chat (with yourself)
7. To leave the chat, either type ".exit", "/exit" or "--exit", or press Ctrl+C, or simply close the terminal


## Description
In this assignment you will design and implement Chit Chat, a distributed chat service where participants can join, exchange messages, and leave the conversation at any time. Chit Chat is a lively playground for exploring the essence of distributed systems: communication, coordination, and the ordering of events in a world without a single shared clock.


## System specs
- S1. Chit Chat is a distributed service that enables clients to exchange chat messages using gRPC for all communication. Students must design the gRPC API, including all service methods and message types.

- S2. The Chit Chat system must follow a distributed topology consisting of one service process and multiple client processes. Each client runs as an independent process that communicates with the service via gRPC. A minimal configuration must include one service instance and at least three concurrently active clients.

- S3. Each participant can publish a valid chat message at any time. A valid message is a UTF-8 encoded string with a maximum length of 128 characters. Publishing is performed through a gRPC call to the Chit Chat service.

- S4. The Chit Chat service must broadcast each published message to all currently active participants. Each broadcast must include the message content and a logical timestamp.

- S5. Participants may join the system at any time.  When a new participant X joins, the service must broadcast a message of the form: "Participant X joined Chit Chat at logical time L". This message must be delivered to all participants, including the newly joined one.

- S6. Participants may leave the system at any time.  When a participant X leaves, the service must broadcast a message of the form: "Participant X left Chit Chat at logical time L". This message must be delivered to all remaining participants.

- S7. When a participant receives any broadcast message, it must: i) display the message content and its logical timestamp on the client interface; and ii) log the message content and its logical timestamp.


## Technical requirements
- [x] The system must be implemented in Go

- [x] The gRPC framework must be used for client–server communication, using Protocol Buffers (you must provide a .proto file) for message definitions.

- [x] The Go log standard library must be used for structured logging of events (e.g. client join/leave notifications, message delivery, and server startup/shutdown).

- [x] Concurrency (local to the server or the clients) must be handled using Go routines and channels for synchronised communication between components

- [x] Every client and the server must be deployed as separate processes

- [x] Each client connection must be served by a dedicated goroutine managed by the server.

- [x] The system must support multiple concurrent client connections without blocking message delivery.

- [x] The system must log the following events:
  * Server startup and shutdown
  * Client connection and disconnection events
  * Broadcast of join/leave messages
  * Message delivery events

- [x] Log messages must include:
  * Timestamp
  * Component name (Server/Client)
  * Event type
  * Relevant identifiers (e.g. Client ID).

- [x] The system can be started with at least three (3) nodes (two client and a server) and it must be able to handle "join" of at least one client and "leave" of at least one client


## Hand-in Requirements
- [x] You must hand in a report (single pdf file) via LearnIT

- [x] You must provide a link to a Git repo with your source code in the report

- [x] Your repo must include a README.md file that describes how to run your program.

- [x] You repo must structured as follows:
> project-root/
>> client/ # contains the client code
>
>> grpc/ # contains .proto file
>
>> server/ # contains the server code
>
>> readme.md  # readme file

- [x] You must include system logs that document the requirements are met, in both the appendix of your report and your repo


## Report requirements
- [x] Discuss, whether you are going to use server-side streaming, client-side streaming, or bidirectional streaming?

- [x] Describe your system architecture - do you have a server-client architecture, peer-to-peer, or something else?

- [x] Describe what RPC methods are implemented, of what type, and what messages types are used for communication

- [x] Describe how you have implemented the calculation of the Lamport timestamps

- [x] Provide a diagram, that traces a sequence of RPC calls together with the Lamport timestamps, that corresponds to a chosen sequence of interactions: Client X joins, Client X Publishes, ..., Client X leaves. 

- [x] You must include system logs that document the requirements are met, in both the appendix of your report and your repo
