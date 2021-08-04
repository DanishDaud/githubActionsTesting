package service

import (
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"os"
)

type Email struct {
	key string
}

type EmailOptions struct {
	FromName    string
	FromEmail   string
	Subject     string
	ToName      string
	ToEmail     string
	TextContent string
	HtmlContent string
}

func NewEmailService(key string) *Email {
	if key == "" {
		panic("sendgrid key is missing")
	}
	return &Email{key: key}
}

func (email *Email) SendEmail(options *EmailOptions) {
	// Sender data.
	emailHost := os.Getenv("SMTP_EMAIL_HOST")

	emailFrom := os.Getenv("SMTP_EMAIL_FROM")

	emailPassword := "Y1k5RjS4S%p#"

	emailPort := os.Getenv("SMTP_PORT")

	logrus.Errorln(emailHost)
	logrus.Errorln(emailFrom)
	logrus.Errorln(emailPassword)
	logrus.Errorln(emailPort)

	m := gomail.NewMessage()
	m.SetHeader("From", emailFrom)
	m.SetHeader("To", options.ToEmail)
	m.SetAddressHeader("Cc", "afzalabbasi.019@gmail.com", "abbasi")
	m.SetHeader("Subject", options.Subject)
	m.SetBody("text/html", options.HtmlContent)
	d := gomail.NewDialer(emailHost, 465, emailFrom, emailPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		logrus.Errorln("Email Send Failed => To Email : ", options.ToEmail, "::: From Email : ", options.FromEmail)
		logrus.Errorln(err)
	}

	logrus.Println("send email Successfully")

}
