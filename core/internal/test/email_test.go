package test

import (
	"cloud_disk/core/internal/define"
	"crypto/tls"
	"github.com/jordan-wright/email"
	"net/smtp"
	"testing"
)

// 测试邮箱发送验证码
func TestEmail(t *testing.T) {
	e := email.NewEmail()
	e.From = "Jordan Wright <18163688304@163.com>"
	e.To = []string{"386244641@qq.com"}
	e.Subject = "验证码发送测试"
	e.HTML = []byte("你的验证码为：<h1>123456</h1>")
	err := e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth("", "18163688304@163.com", define.MailPassword,
		"smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
	if err != nil {
		t.Fatal(err)
	}

}
