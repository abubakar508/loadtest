package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type Mailer struct {
	Host     string
	Port     string
	Username string
	Password string
	Auth     smtp.Auth
	From     string
}

func NewMailer() *Mailer {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", username, password, host)

	return &Mailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Auth:     auth,
		From:     username,
	}
}

func (m *Mailer) SendEmail(to string, subject string, body string) error {
	addr := m.Host + ":" + m.Port

	headers := make(map[string]string)
	headers["From"] = m.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + body)

	return smtp.SendMail(addr, m.Auth, m.From, []string{to}, []byte(msg.String()))
}
