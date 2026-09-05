output "function_name" {
  value = aws_lambda_function.auth.function_name
}

output "function_arn" {
  value = aws_lambda_function.auth.arn
}

output "invoke_arn" {
  value       = aws_lambda_function.auth.invoke_arn
  description = "Valor de lambda_authorizer_invoke_arn no postech-tc3-infra-k8s"
}

output "log_group" {
  value = aws_cloudwatch_log_group.lambda.name
}
