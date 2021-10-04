package internal

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type MailInfo struct {
	smtpHost string
	smtpPort int
	from     string
	pwd      string
	sendTo   []string
}

func NewMailInfo(smtpHost string, smtpPort int, from, pwd, sendTo string) *MailInfo {
	return &MailInfo{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
		pwd:      pwd,
		sendTo:   strings.Split(sendTo, ","),
	}
}

func (m *MailInfo) sendStackTraceByEmail(stackTrace string) {

	message := []byte("Subject: [Mongoplayground] New server error\r\n" +
		"\r\n" +
		stackTrace + "\r\n")

	err := smtp.SendMail(
		fmt.Sprintf("%v:%d", m.smtpHost, m.smtpPort),
		smtp.PlainAuth("", m.from, m.pwd, m.smtpHost),
		m.from,
		m.sendTo,
		message,
	)
	if err != nil {
		log.Printf("fail to send mail: %v", err)
	}
}
