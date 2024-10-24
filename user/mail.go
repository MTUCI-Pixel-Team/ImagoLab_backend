package user

import (
	"RestAPI/core"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
	"gopkg.in/gomail.v2"
)

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

	msgHTML := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
	    <meta charset="UTF-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	    <title>Account Activation</title>
	    <style>
	        body {
	            font-family: Arial, sans-serif;
	            background-color: #f4f4f4;
	            margin: 0;
	            padding: 0;
	        }
	        .container {
	            width: 100%%;
	            max-width: 600px;
	            margin: 20px auto;
	            background-color: #ffffff;
	            padding: 20px;
	            border-radius: 8px;
	            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
	        }
	        .header {
	            text-align: center;
	            margin-bottom: 20px;
	        }
	        .content {
	            font-size: 16px;
	            line-height: 1.5;
	            color: #333333;
	        }
	        .activation-code {
	            font-size: 24px;
	            font-weight: bold;
	            color: #4CAF50;
	            text-align: center;
	            margin: 20px 0;
	        }
	        .footer {
	            font-size: 12px;
	            color: #888888;
	            text-align: center;
	            margin-top: 20px;
	        }
	    </style>
	</head>
	<body>
	    <div class="container">
	        <div class="header">
	            <h2>Account Activation</h2>
	        </div>
	        <div class="content">
	            <p>Hello,</p>
	            <p>Thank you for registering. To activate your account, please use the following confirmation code:</p>
	            <div class="activation-code">%d</div>
	            <p>If you did not request this email, please ignore it.</p>
	        </div>
	        <div class="footer">
	            <p>&copy; 2024 ImagoLab. All rights reserved.</p>
	        </div>
	    </div>
	</body>
	</html>
	`, activationCode)

	m.SetBody("text/html", msgHTML)

	return d.DialAndSend(m)
}

func generateActivationCode() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Intn(99999) + 100000
}
