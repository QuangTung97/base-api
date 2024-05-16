package main

import (
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	file, err := os.OpenFile("/tmp/gin-lock-db", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()

	log.Println("LOCK STARTED")
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		panic(err)
	}
	log.Println("LOCK GRANTED")

	time.Sleep(10 * time.Second)
	log.Println("LOCK ENDED")
}
