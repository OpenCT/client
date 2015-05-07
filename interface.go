package main

import "github.com/tarm/serial"


type Connection struct{
  port *Port
  i int
  imageBuffer []byte
}
function (c *Connection) Write(a []byte){
  c.write(a)
}
function (c *Connection) Read() chan int{
  done := make(chan int)
  go func(){
    //loop
    //done <- 0
  }
}
