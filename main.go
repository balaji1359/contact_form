package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func sendEmail(name, email, message string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	to := []string{smtpEmail}
	subject := "New Contact Form Submission"
	body := fmt.Sprintf("Name: %s\nEmail: %s\nMessage:\n%s", name, email, message)
	msg := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, to, msg)
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var form ContactForm
	err := json.Unmarshal([]byte(request.Body), &form)
	if err != nil {
		log.Printf("Failed to parse request: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid form data"}, nil
	}

	if form.Name == "" || form.Email == "" || form.Message == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "All fields are required"}, nil
	}

	err = sendEmail(form.Name, form.Email, form.Message)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to send email"}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Message sent successfully!"}, nil
}

func main() {
	if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
		lambda.Start(lambdaHandler)
	} else {
		http.HandleFunc("/send-email", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
				return
			}

			var form ContactForm
			err := json.NewDecoder(r.Body).Decode(&form)
			if err != nil || form.Name == "" || form.Email == "" || form.Message == "" {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			err = sendEmail(form.Name, form.Email, form.Message)
			if err != nil {
				http.Error(w, "Failed to send email", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Message sent successfully!"))
		})
		fmt.Println("Running locally at http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
