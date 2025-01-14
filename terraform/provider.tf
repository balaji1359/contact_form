provider "aws" {
  region = "ap-south-1"
}

terraform {
  backend "s3" {
    bucket = "balaji-nagisetty"
    key    = "contact_form.tfstate"
    region = "us-east-1"
  }
}