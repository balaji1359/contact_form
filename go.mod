module contact_form_lambda

go 1.23

require github.com/aws/aws-lambda-go v1.47.0 // indirect



curl -X POST https://27yxseq1z8.execute-api.ap-south-1.amazonaws.com/prod/send-email \
  -H "Content-Type: application/json" \
  -d '{"subject": "Test Email", "body": "This is a test email sent from my Go app!"}'
