[![Build Status](https://cloud.drone.io/api/badges/elahe-dastan/P2P/status.svg)](https://cloud.drone.io/elahe-dastan/P2P)

# P2P

I want to practice socket programming in this repository

## What the project does

This project is a simulation of peer to peer applications, each client has a list of IPs which is its cluster then <br/>
each node send its list to all the nodes in the list so they can update their lists, each node can also ask for a file<br/>
, and get the file from the first node that answers. <br/>

## Protocols

One of the most important things I have to do in this project is to define a good application protocol for <br/>
communication, for example the udp server gets two type of messages one for requesting a file, the other for giving list,<br/>
the format of the message should divide these two from each other.

| UDP Server             | Example                                   |
|:-----------------------| ------------------------------------------|
| Discover,"ip addresses | "Discover,127.0.0.1:1373,127.0.0.1:1378"  |
| Get,file name          | "Get,resume.pdf                           |
| File,tcp port          | "File,33680"                              |