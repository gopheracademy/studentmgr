package exec

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bketelsen/studentmgr/models"
	"github.com/kr/pretty"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type ExecManager struct {
	ScriptPath string
}

func (em *ExecManager) Exists(username string) (bool, error) {
	path := filepath.Join("/", "home", username)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err

}
func (em *ExecManager) NewStudent(student *models.Student) error {

	pretty.Println(student)

	script := filepath.Join(em.ScriptPath, "user", "create.sh")
	cmd := exec.Command(script, student.Username, student.Password, student.Email)

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	u, err := user.Lookup(student.Username)
	if err != nil {
		fmt.Println(err)
		return err
	}
	student.UID, _ = strconv.Atoi(u.Uid)
	student.HomeDirectory = "/home/" + u.Username
	fmt.Printf("User: %s  ID: %s\n", u.Username, u.Uid)

	wide := filepath.Join(em.ScriptPath, "user", "wide")
	t, err := template.ParseFiles(wide)
	if err != nil {
		fmt.Print(err)
		return err
	}

	widepath := filepath.Join("/", "home", u.Username, ".newuser", "wide.sh")

	f, err := os.Create(widepath)
	if err != nil {
		fmt.Println("create file: ", err)
		return err
	}
	defer f.Close()
	config := map[string]string{"Password": student.Password, "Email": student.Email}
	err = t.Execute(f, config)
	if err != nil {
		fmt.Print("execute: ", err)
		return err
	}

	wideservice := filepath.Join(em.ScriptPath, "user", "wide.service")
	wt, err := template.ParseFiles(wideservice)
	if err != nil {
		fmt.Print(err)
		return err
	}

	unitpath := filepath.Join("/", "home", u.Username, ".config", "systemd", "user")
	err = os.MkdirAll(unitpath, 0755)
	if err != nil {
		fmt.Print(err)
		return err
	}
	unitpath = filepath.Join("/", "home", u.Username, ".config", "systemd", "user", "wide.service")

	sf, err := os.Create(unitpath)
	if err != nil {
		fmt.Println("create unit file: ", err)
		return err
	}
	defer sf.Close()
	uconfig := map[string]string{"Username": u.Username, "UID": u.Uid}
	err = wt.Execute(sf, uconfig)
	if err != nil {
		fmt.Print("execute: ", err)
		return err
	}

	err = os.Chmod(widepath, 0755)
	if err != nil {
		fmt.Print("chmod: ", err)
		return err
	}

	template := filepath.Join(em.ScriptPath, "user", "template")
	read, err := ioutil.ReadFile(template)
	if err != nil {
		fmt.Println(err)
		return err
	}

	newContents := strings.Replace(string(read), "UID", u.Uid, -1)

	newContents = strings.Replace(string(newContents), "USERNAME", u.Username, -1)

	fmt.Println(newContents)

	newpath := filepath.Join("/opt/caddy/sites", u.Username)

	err = ioutil.WriteFile(newpath, []byte(newContents), 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	script = filepath.Join(em.ScriptPath, "user", "caddy.sh")
	cmd = exec.Command(script)

	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("created student")

	for _, c := range student.Courses {
		fmt.Println("creating course", c, c.ID, strconv.Itoa(int(c.ID)))

		script := filepath.Join(em.ScriptPath, "classes", strconv.Itoa(int(c.ID)), "prepare.sh")
		cmd := exec.Command(script, student.Username, student.Password)

		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			return err
		}

	}
	fmt.Println("prepared course")

	return nil

}

func notify(email, username string) {

	post := `{"email_address": "EMAIL","status": "subscribed"}`
	post = strings.Replace(post, "EMAIL", email, -1)

	mcurl := "https://us1.api.mailchimp.com/3.0/lists/dccb0487a6/members"

	req, err := http.NewRequest("POST", mcurl, bytes.NewBuffer([]byte(post)))
	req.Header.Set("Authorization", "apikey 669fb1a40d1afbe12d67f24f3ba504dc-us1")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("MC Error: ", err)
	}

	fmt.Println("MC response Status:", resp.Status)
	fmt.Println("MC response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("MC response Body:", string(body))

	// email finished thing

	from := mail.NewEmail("Brian Ketelsen", "me@brianketelsen.com")
	to := mail.NewEmail(username, email)

	subject := "Your account is ready to use for your class!"

	message := "Congratulations, your virtual student classroom account is ready to use. \nIf prompted for a username and password, use the username/password combination you \ncreated when you enrolled. \n\n\n"

	message = message + "Welcome.  Your access information is below: \n\n"
	ssh := fmt.Sprintf("ssh %s@students.brianketelsen.com", username)
	ide := fmt.Sprintf("https://ide.brianketelsen.com/%s/", username)
	shell := fmt.Sprintf("https://shell.brianketelsen.com/")
	message = message + "SSH Access: \n "
	message = message + ssh + "\n"
	message = message + "Web Shell: \n "
	message = message + shell + "\n"
	message = message + "Web IDE - ONLY available after your first login: \n "
	message = message + ide + "\n\n"

	message = message + "\n\n Using these servers is governed under the terms and conditions of your course. \nThe servers will be decomissioned at the end of your class, and are solely \nfor your private use during the duration of the course.\n\n Thanks!\n\n Brian Ketelsen "
	time.Sleep(45 * time.Second)
	content := mail.NewContent("text/plain", message)
	m := mail.NewV3MailInit(from, subject, to, content)

	request := sendgrid.GetRequest("L9dMYDKFQGyg9A4rG8GWng", "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

}
