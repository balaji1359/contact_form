package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"net/smtp"
)

type EmailRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func sendEmail(subject, body string) error {
	// Get environment variables for email credentials
	email := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Compose message
	to := []string{email}
	message := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))

	// Authenticate and send the email
	auth := smtp.PlainAuth("", email, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, email, to, message)
	return err
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var emailReq EmailRequest
	err := json.NewDecoder(r.Body).Decode(&emailReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if emailReq.Subject == "" || emailReq.Body == "" {
		http.Error(w, "Subject and body are required", http.StatusBadRequest)
		return
	}

	err = sendEmail(emailReq.Subject, emailReq.Body)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully!"))
}

func main() {
	http.HandleFunc("/send-email", emailHandler)
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
