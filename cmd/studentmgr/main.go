package main

import (
	"fmt"
	"log"
	"net"

	h "net/http"

	"github.com/bketelsen/studentmgr/exec"
	"github.com/bketelsen/studentmgr/http"
	"github.com/bketelsen/studentmgr/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

var sd string // script directory

func main() {

	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	server := http.NewServer(db)
	db.AutoMigrate(&models.Student{}, &models.Course{})
	db.LogMode(true)

	// Initalize
	Admin := admin.New(&qor.Config{DB: db})

	// Create resources from GORM-backend model
	Admin.AddResource(&models.Student{})
	Admin.AddResource(&models.Course{})

	// Register route
	mux := h.NewServeMux()
	// amount to /admin, so visit `/admin` to view the admin interface
	Admin.MountTo("/admin", mux)
	fmt.Println("Admin Listening on: 9000")
	go h.ListenAndServe(":9000", mux)
	execmgr := exec.ExecManager{
		ScriptPath: "/home/bketelsen/src/github.com/bketelsen/studentmgr/studentmgrscripts",
	}
	server.ExecManager = &execmgr
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Panic(err)
	}
	log.Println(server.Serve(ln))

}
