package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain     string
	Host       string
	Port       int
	Username   string
	Password   string
	Encryption string
	FromAdress string
	FromName   string
	WaitGroup  *sync.WaitGroup
	MailerChan chan Message
	ErrorChan  chan error
	DoneChan   chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

func (app *Config) listenForMail() {
	for {
		select {
		case msg := <-app.Mailer.MailerChan:
			go app.Mailer.sendMail(msg, app.Mailer.ErrorChan)
		case err := <-app.Mailer.ErrorChan:
			app.ErrorLog.Println(err)
		case <-app.Mailer.DoneChan:
			return
		}
	}
}

func (m *Mail) sendMail(message Message, errorChan chan error) {
	defer m.WaitGroup.Done()

	if message.Template == "" {
		message.Template = "mail"
	}

	if message.From == "" {
		message.From = m.FromAdress
	}

	if message.FromName == "" {
		message.FromName = m.FromName
	}

	data := map[string]any{
		"message": message.Data,
	}

	message.DataMap = data

	formatedMesage, err := m.buildHTMLMessage(message)
	if err != nil {
		errorChan <- err
		return
	}

	plainMessage, err := m.buildPlainTextMessage(message)
	if err != nil {
		errorChan <- err
		return
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(message.From).AddTo(message.To).SetSubject(message.Subject)

	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formatedMesage)

	if len(message.Attachments) > 0 {
		for _, x := range message.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(message Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", message.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()

	formattedMessage, err = m.inlinceCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(message Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", message.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) inlinceCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
