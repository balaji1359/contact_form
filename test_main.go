package main

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-lambda-go/events"
)

func main() {
    request := events.APIGatewayProxyRequest{
        Body: `{"subject": "Test Local Email", "body": "This is a local test."}`,
    }

    response, err := handler(context.Background(), request)
    if err != nil {
        log.Fatalf("Error executing handler: %v", err)
    }

    fmt.Printf("Response: %+v\n", response)
}
