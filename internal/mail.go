// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2017 Adrien Petel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
