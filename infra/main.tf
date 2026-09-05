locals {
  name = "${var.project_name}-${var.environment}-auth"

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
}

resource "aws_security_group" "lambda" {
  name        = "${local.name}-sg"
  description = "Egress da lambda de autenticacao para o RDS"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.tags
}

resource "aws_cloudwatch_log_group" "issuer" {
  name              = "/aws/lambda/${local.name}-issuer"
  retention_in_days = var.log_retention_days

  tags = local.tags
}

resource "aws_cloudwatch_log_group" "authorizer" {
  name              = "/aws/lambda/${local.name}-authorizer"
  retention_in_days = var.log_retention_days

  tags = local.tags
}

resource "aws_lambda_function" "issuer" {
  function_name = "${local.name}-issuer"
  role          = data.aws_iam_role.lab.arn

  filename         = var.issuer_artifact_path
  source_code_hash = filebase64sha256(var.issuer_artifact_path)

  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]
  timeout       = 15
  memory_size   = 256

  vpc_config {
    subnet_ids         = data.aws_subnets.default.ids
    security_group_ids = [aws_security_group.lambda.id]
  }

  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
      JWT_TTL      = var.jwt_ttl
    }
  }

  depends_on = [aws_cloudwatch_log_group.issuer]

  tags = local.tags
}

resource "aws_lambda_function" "authorizer" {
  function_name = "${local.name}-authorizer"
  role          = data.aws_iam_role.lab.arn

  filename         = var.authorizer_artifact_path
  source_code_hash = filebase64sha256(var.authorizer_artifact_path)

  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]
  timeout       = 5
  memory_size   = 128

  environment {
    variables = {
      JWT_SECRET = var.jwt_secret
    }
  }

  depends_on = [aws_cloudwatch_log_group.authorizer]

  tags = local.tags
}
