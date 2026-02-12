package utils

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendOTPEmail sends a 6-digit OTP code to the specified email address
func SendOTPEmail(to, code string) error {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	// Fallback for development if not set
	if host == "" {
		log.Println("WARNING: SMTP_HOST not set. Email not sent. OTP Code:", code)
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587 // Default SMTP port
	}

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your Password Reset OTP")
	m.SetBody("text/html", "<h2>Password Reset</h2><p>Your 6-digit OTP code is: <b>"+code+"</b></p><p>This code will expire in 5 minutes.</p>")

	d := gomail.NewDialer(host, port, user, pass)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
