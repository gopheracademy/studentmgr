package http

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"html/template"

	"github.com/bketelsen/studentmgr/exec"
	"github.com/bketelsen/studentmgr/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type SignupResponse struct {
	Error        bool
	ErrorMessage string
	ShellURL     string
	IDEURL       string
	Username     string
}

type Server struct {
	Router      *mux.Router
	db          *gorm.DB
	ExecManager *exec.ExecManager
}

func NewServer(db *gorm.DB) *Server {
	s := &Server{}
	s.db = db
	router := mux.NewRouter()
	s.Router = router
	router.HandleFunc("/students/{id}", s.GetStudent).Methods("GET")
	router.HandleFunc("/enroll", s.ServeEnroll).Methods("GET")
	router.HandleFunc("/enroll", s.Enroll).Methods("POST")
	return s
}

func (s *Server) Serve(l net.Listener) error {
	return http.Serve(l, s.Router)
}

func (s *Server) ServeEnroll(w http.ResponseWriter, req *http.Request) {
	t, err := template.New("body").Parse(form)
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	err = t.Execute(w, &SignupResponse{})
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

func (s *Server) Enroll(w http.ResponseWriter, req *http.Request) {

	t, err := template.New("body").Parse(form)
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	wr := &SignupResponse{}

	courseid := req.PostFormValue("coursetoken")
	if courseid == "" {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: "Course Token Required"})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}
	var course models.Course
	err = s.db.Where("id=?", courseid).First(&course).Error
	if err != nil {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: "Bad Course Token"})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}

	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	password2 := req.PostFormValue("password2")
	fmt.Println(password, password2)
	if password != password2 {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: "Passwords must match"})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}

	name := req.PostFormValue("name")
	email := req.PostFormValue("email")

	// check for homedir here and fail if username is taken

	exists, err := s.ExecManager.Exists(username)

	if err != nil {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: err.Error()})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}

	if exists {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: "Username Taken"})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return

	}

	student := &models.Student{
		Username: username,
		Password: password,
		FullName: name,
		Email:    email,
		Courses:  []models.Course{course},
	}
	err = s.db.Create(student).Error
	if err != nil {
		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: err.Error()})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}

	fmt.Println(student)

	err = s.ExecManager.NewStudent(student)
	if err != nil {

		w.WriteHeader(400)
		e := t.Execute(w, &SignupResponse{Error: true, ErrorMessage: err.Error()})
		if e != nil {
			log.Print("template executing error: ", err)
		}
		return
	}
	w.WriteHeader(201)
	wr.Username = student.Username
	wr.ShellURL = "https://shell.brianketelsen.com"
	wr.IDEURL = fmt.Sprintf("https://students.brianketelsen.com/%s", student.Username)
	e := t.Execute(w, wr)
	if e != nil {
		log.Print("template executing error: ", err)
	}

}
func (s *Server) GetStudent(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, err := strconv.Atoi(params["id"])
	fmt.Println(id)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error"))

	}
	fmt.Println(id)
	var student models.Student
	err = s.db.Where("id=?", id).First(&student).Error
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}

	w.Write([]byte(fmt.Sprintf("%v", student)))
}

const form = `<!DOCTYPE html>
<html>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/font-awesome/4.3.0/css/font-awesome.min.css">
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
<link href='http://fonts.googleapis.com/css?family=Varela+Round' rel='stylesheet' type='text/css'>
<script   src="http://code.jquery.com/jquery-3.1.1.min.js"   integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8="   crossorigin="anonymous"></script>
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />
<body>

<div class="container">
{{ if .Error }}
<div class="alert alert-danger">
<strong>Error:<strong> {{.ErrorMessage}}
</div>
{{ end  }}
{{ if .Username}}
<div class="alert alert-success">
  <strong>Created<strong> {{.Username}} was created<br/>
  <strong>Shell Access:<strong> <a target=new href="{{.ShellURL}}">Click Here</a><br/>
  <strong>Web IDE:<strong> <a target=new href="{{.IDEURL}}">Click Here</a> (available after first login)
</div>
{{ end  }}
        <div class="row centered-form">
        <div class="col-xs-12 col-sm-8 col-md-4 col-sm-offset-2 col-md-offset-4">
        	<div class="panel panel-default">
        		<div class="panel-heading">
				<h3 class="panel-title">Enter your user information to join:</h3>
			 			</div>
			 			<div class="panel-body">
			    		<form role="form" action="/enroll" method="POST">
			    					<div class="form-group">
										<input type="text" name="username" id="username" class="form-control input-sm" placeholder="username">
			    					</div>
			    					<div class="form-group">
			    						<input type="text" name="name" id="name" class="form-control input-sm" placeholder="Full Name">
			    					</div>

			    			<div class="form-group">
			    				<input type="email" name="email" id="email" class="form-control input-sm" placeholder="Email Address">
			    			</div>

			    			<div class="row">
			    				<div class="col-xs-6 col-sm-6 col-md-6">
			    					<div class="form-group">
			    						<input type="password" name="password" id="password" class="form-control input-sm" placeholder="Password">
			    					</div>
			    				</div>
			    				<div class="col-xs-6 col-sm-6 col-md-6">
			    					<div class="form-group">
			    						<input type="password" name="password2" id="password2" class="form-control input-sm" placeholder="Confirm Password">
			    					</div>
			    				</div>
			    			</div>
			    			<div class="form-group">
			    				<input type="text" name="coursetoken" id="coursetoken" class="form-control input-sm" placeholder="CourseToken">
			    			</div>
			    			
			    			<input type="submit" value="Register" class="btn btn-info btn-block">
			    		
			    		</form>
			    	</div>
	    		</div>
    		</div>
    	</div>
    </div>
</body>
</html>`
