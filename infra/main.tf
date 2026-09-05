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

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${local.name}"
  retention_in_days = var.log_retention_days

  tags = local.tags
}

resource "aws_lambda_function" "auth" {
  function_name = local.name
  role          = data.aws_iam_role.lab.arn

  filename         = var.artifact_path
  source_code_hash = filebase64sha256(var.artifact_path)

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

  depends_on = [aws_cloudwatch_log_group.lambda]

  tags = local.tags
}
