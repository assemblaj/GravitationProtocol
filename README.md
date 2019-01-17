# Gravitation Protocol 
Gravitation Protocol provides a way to find peers with simlar properties, functions, and goals. 

## How it works: 
- Local peers send a "gravitation" request which includes a Metadata ID (list of properties that describe the peer) and their "orbit" (a collection of peers they've collected and their properties). 
- Remote peers respond with the same information 
- Upon recieving a gravitation request or response, both peers determine whether to accept a peer into their orbit based on their Metadata ID, and their orbits, as well as performing this process on the remote peers' orbit as well. 

## How to use: 
Clone the project
make deps 
go build 


./GravitationProtocol -listen /ip4/127.0.0.1/tcp/{port}         : Runs Gravitation with specific multiadresss and default profile
./GravitationProtocol -listen /ip4/127.0.0.1/tcp/{another port} : To create other peer 


## Relavant Flags:
- save [filename]  : Saves Gravitation Data to file after program is closed [Ctrl+C] 
- load [filename] : Loads gravitation data from file 
- profile  "test1 test2 test3 test4"  : Double-quoted, space separated list of profile values. 
- test [filename]:  Runs test with specified test file
