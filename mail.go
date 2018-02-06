package main

import (
	"log"
	"net/smtp"
	"time"
)

type mail struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Host     string        `json:"host"`
	Port     string        `json:"port"`
	From     string        `json:"from"`
	To       []string      `json:"to"`
	Subject  string        `json:"subject"`
	Content  string        `json:"content"`
	Duration time.Duration `json:"duration"`
}

func (p *mail) sendEmail() {
	count := 0
	// Set up authentication information.
	auth := smtp.PlainAuth("", p.Username, p.Password, p.Host)
	// Connect to the server, authenticate, set the sender and recipient, and send the email all in one step.
	msgStr := "From: " + p.From + "\r\n"
	for _, to := range p.To {
		msgStr += "To: " + to + "\r\n"
	}
	msgStr += "Subject: " + p.Subject + "\r\n" +
		"\r\n" +
		p.Content + "\r\n"

	err := smtp.SendMail(p.Host+":"+p.Port, auth, p.Username, p.To, []byte(msgStr))
	if err != nil {
		log.Println(err)
		for {
			err := smtp.SendMail(p.Host+":"+p.Port, auth, p.Username, p.To, []byte(msgStr))
			if err == nil {
				break
			}
			log.Println(err)
			//如果邮件发送失败的重试次数超过十次就放弃发送
			if count++; count > 10 {
				break
			}
			time.Sleep(p.Duration)
		}
	}
}
