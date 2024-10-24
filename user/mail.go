package user

import (
	"RestAPI/core"
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	ActivationCode int
	ResetCode      string
}

func SendActivationEmail(toEmail string, activationCode int) error {
	if toEmail == "" {
		return fmt.Errorf("Email is empty")
	}
	if activationCode > 999999 || activationCode < 100000 {
		return fmt.Errorf("Invalid activation code")
	}

	d := gomail.NewDialer(core.MAIL_HOST, core.MAIL_PORT, core.MAIL_USER, core.MAIL_PASSWORD)
	d.SSL = true

	m := gomail.NewMessage()
	m.SetHeader("From", core.MAIL_USER)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Account Activation")

	data := EmailData{
		ActivationCode: activationCode,
	}

	var buf bytes.Buffer

	if err := core.ACTIVATE_EMAIL_TEMPLATE.Execute(&buf, data); err != nil {
		return err
	}

	msgHTML := buf.String()

	m.SetBody("text/html", msgHTML)

	return d.DialAndSend(m)
}

func SendResetPasswordEmail(toEmail, resetToken string) error {
	if toEmail == "" {
		return fmt.Errorf("Email is empty")
	}
	if resetToken == "" {
		return fmt.Errorf("Invalid reset code")
	}

	d := gomail.NewDialer(core.MAIL_HOST, core.MAIL_PORT, core.MAIL_USER, core.MAIL_PASSWORD)
	d.SSL = true

	m := gomail.NewMessage()
	m.SetHeader("From", core.MAIL_USER)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Reset Password")

	data := EmailData{
		ResetCode: resetToken,
	}

	var buf bytes.Buffer

	if err := core.RESET_PASSWORD_TEMPLATE.Execute(&buf, data); err != nil {
		return err
	}

	msgHTML := buf.String()

	m.SetBody("text/html", msgHTML)

	return d.DialAndSend(m)
}

func generateActivationCode() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Intn(99999) + 100000
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
