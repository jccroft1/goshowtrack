package mail

import (
	"fmt"
	"log"
	"strconv"

	gomail "gopkg.in/gomail.v2"
)

var (
	server   string
	port     int
	username string
	password string
	from     string
)

func Setup(_server string, _port string, _username string, _password string, _from string) error {
	if server != "" {
		log.Println("No server provided, SMTP disabled. ")
		return nil
	}
	server = _server
	var err error
	port, err = strconv.Atoi(_port)
	if err != nil {
		return fmt.Errorf("invalid port provided: %s", _port)
	}
	username = _username
	password = _password
	from = _from

	return nil
}

func SendMagicLink(email, token string) error {
	link := fmt.Sprintf("http://localhost:8080/login?token=%s", token)

	fmt.Println("Login link", link)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Magic Login Link")
	m.SetBody("text/plain", "Click to sign in: "+link)

	if server == "" {
		return nil
	}

	d := gomail.NewDialer(server, port, username, password)
	return d.DialAndSend(m)
}
