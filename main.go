package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string           `json:"body"`
}

func sendEmail(name, email, message string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Create email content
	to := []string{smtpEmail}
	subject := "New Contact Form Submission"
	emailBody := fmt.Sprintf("From: %s\nEmail: %s\nMessage:\n%s", name, email, message)
	
	// Format email message
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("To: %s\nSubject: %s\n%s\n%s", 
		smtpEmail, subject, mime, emailBody))

	// Setup authentication
	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, to, msg)
	if err != nil {
		log.Printf("SMTP error: %v", err)
		return err
	}

	return nil
}

func createResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: body,
	}
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Log the incoming request
	log.Printf("Processing request: %s", request.Body)

	// Parse the request body
	var form ContactForm
	if err := json.Unmarshal([]byte(request.Body), &form); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return createResponse(400, `{"message": "Invalid request format"}`), nil
	}

	// Validate required fields
	if form.Name == "" || form.Email == "" || form.Message == "" {
		return createResponse(400, `{"message": "All fields are required"}`), nil
	}

	// Send email
	if err := sendEmail(form.Name, form.Email, form.Message); err != nil {
		log.Printf("Error sending email: %v", err)
		return createResponse(500, `{"message": "Failed to send email"}`), nil
	}

	return createResponse(200, `{"message": "Message sent successfully"}`), nil
}

func main() {
	lambda.Start(handleRequest)
}