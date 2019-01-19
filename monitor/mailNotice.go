package monitor

import (
	"strings"
	"net/smtp"
)

var Mail *MailInfo

type MailInfo struct {
	MailUser    string
	MailPasswd  string
	SmtpHost    string
	ReceiveMail []string
}

func NewMail(user, passwd, smtpHost string, receiveMail []string) (*MailInfo) {
	if Mail != nil {
		return Mail
	}
	Mail = &MailInfo{user, passwd, smtpHost, receiveMail}
	return Mail
}

func (mailInfo *MailInfo) sendMail(content string) {
	// 邮箱账号
	user := mailInfo.MailUser
	//注意，此处为授权码、不是密码
	password := mailInfo.MailPasswd
	//smtp地址及端口
	host := mailInfo.SmtpHost
	//接收者，内容可重复，邮箱之间用；隔开
	var to_remail string
	for _, remail := range mailInfo.ReceiveMail {
		to_remail += ";" + remail
	}
	//邮件主题
	subject := "Lancet监控容器挂掉"
	hp := strings.Split(mailInfo.SmtpHost, ":")

	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	content_type = "Content-Type: text/plain" + "; charset=UTF-8"

	msg := []byte("To: " + to_remail + "\r\nFrom: " + user + "<" + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + content)
	logger.Debugf("mailInfo  is [%v]!", mailInfo)
	logger.Debugf("sendMail info is [%s]!", content)
	err := smtp.SendMail(host, auth, user, mailInfo.ReceiveMail, msg)
	if err != nil {
		logger.Errorf("Send mailInfo fail! Error is %s", err.Error())
	} else {
		logger.Debugf("Send mailInfo success!")
	}
}
