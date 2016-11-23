package main

import (
	"log"

	"github.com/bketelsen/studentmgr/bolt"
	"github.com/kr/pretty"
)

var sd string // script directory

func main() {
	client := bolt.NewClient()
	client.Path = "/home/bketelsen/studentmgr.db"
	err := client.Open()
	if err != nil {
		log.Panic(err)
	}

	session := client.Connect()
	ss, err := session.StudentStorage().List()

	pretty.Println(ss)

}
