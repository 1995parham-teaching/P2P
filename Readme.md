[![Build Status](https://cloud.drone.io/api/badges/elahe-dastan/P2P/status.svg)](https://cloud.drone.io/elahe-dastan/P2P)
#P2P
I want to practice socket programming in this repository

##Protocols
One of the most important things we have to do in this project is to define a good application protocol for <br/>
communication, for example the udp server gets two type of messages one for requesting list the other for giving list,<br/>
the format of the message should divide this two from each other

| UDP Client          | UDP Server         | Example                                                          |
| ------------------- |:-------------------| -----------------------------------------------------------------|
| get                 | list,"ip addresses | client:"get"<br/>server:"list,127.0.0.1                          |
| List                | List,id1-id2,...   | client:"List"<br/>server:"List,2-3-4                             |
| Send,id-id-...,body | Send,body          | client:"Send,2-3,I will be late<br/>server:"Send,I will be late" |