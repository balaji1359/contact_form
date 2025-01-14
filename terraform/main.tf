# Variables
variable "smtp_email" {
  description = "SMTP email address"
  type        = string
}

variable "smtp_password" {
  description = "SMTP password (app password)"
  type        = string
  sensitive   = true
}

# IAM Role for Lambda execution
resource "aws_iam_role" "lambda_exec_role" {
  name = "contact_form_lambda_exec_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# Attach the basic Lambda execution policy
resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Inline policy for KMS Decrypt permission
resource "aws_iam_role_policy" "lambda_kms_policy" {
  name = "lambda_kms_decrypt_policy"
  role = aws_iam_role.lambda_exec_role.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "kms:Decrypt"
        Resource = "arn:aws:kms:ap-south-1:880622142287:key/eef2be25-a387-4889-bc86-bf7543b2c647"
      }
    ]
  })
}

# Lambda function configuration
resource "aws_lambda_function" "email_lambda" {
  filename         = "lambda.zip"
  function_name    = "email_handler"
  role            = aws_iam_role.lambda_exec_role.arn
  handler         = "bootstrap"  # Changed to bootstrap for provided.al2023
  runtime         = "provided.al2023"
  architectures   = ["x86_64"]
  timeout         = 30
  source_code_hash = filebase64sha256("lambda.zip")

  environment {
    variables = {
      SMTP_EMAIL    = var.smtp_email
      SMTP_PASSWORD = var.smtp_password
    }
  }
}

# API Gateway configuration
resource "aws_apigatewayv2_api" "http_api" {
  name          = "email_api"
  protocol_type = "HTTP"
  cors_configuration {
    allow_headers = ["Content-Type"]
    allow_methods = ["POST"]
    allow_origins = ["*"]
    max_age       = 300
  }
}

resource "aws_apigatewayv2_stage" "api_stage" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 100
    throttling_rate_limit  = 50
  }
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id               = aws_apigatewayv2_api.http_api.id
  integration_type     = "AWS_PROXY"
  integration_method   = "POST"
  integration_uri      = aws_lambda_function.email_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "email_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "POST /send-email"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.email_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

# Outputs
output "api_endpoint" {
  value = "${aws_apigatewayv2_api.http_api.api_endpoint}/send-email"
  description = "API Gateway endpoint URL for sending email"
}