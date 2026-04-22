package test

import (
	"crypto/tls"
	"net/smtp"
	"os"
	"testing"

	"github.com/jordan-wright/email"
)

// 测试邮箱发送验证码（需要 MAIL_PASSWORD 环境变量）
func TestEmail(t *testing.T) {
	password := os.Getenv("MAIL_PASSWORD")
	if password == "" {
		t.Skip("MAIL_PASSWORD not set, skip SMTP test")
	}

	e := email.NewEmail()
	e.From = "Jordan Wright <18163688304@163.com>"
	e.To = []string{"386244641@qq.com"}
	e.Subject = "验证码发送测试"
	e.HTML = []byte("你的验证码为：<h1>123456</h1>")
	err := e.SendWithTLS("smtp.163.com:465",
		smtp.PlainAuth("", "18163688304@163.com", password, "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
	if err != nil {
		t.Fatal(err)
	}
}
