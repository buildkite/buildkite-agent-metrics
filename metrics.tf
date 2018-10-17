/*
 * Lambda function for collecting buildkite metrics and sending them to Cloudwatch
 */

data "aws_caller_identity" "current" {}

terraform {
  required_version = "0.11.7"
}

// AWS region. e.g. "us-east-1"
variable region {
  default = "us-east-1"
}

variable "buildkite_agent_token" {
}

variable "buildkite_queue" {
  default = "default"
}

provider "aws" {
  region = "${var.region}"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "metrics_role" {
  name = "bk_monitor_lambda_role"
  path = "/"
  assume_role_policy = <<POLICY
{
   "Version":"2012-10-17",
   "Statement":[
      {
         "Effect": "Allow",
         "Principal": {
            "Service":"lambda.amazonaws.com"
         },
         "Action":"sts:AssumeRole"
      }
   ]
}
POLICY
}

resource "aws_iam_role_policy" "metrics_lambda_policy" {
  name = "metrics_lambda_policy"
  role = "${aws_iam_role.metrics_role.id}"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "cloudwatch:PutMetricData"
                ],
            "Resource": "*"
        }
    ]
}
POLICY
}

resource "aws_lambda_function" "buildkite-agent-metrics-function" {
  function_name = "buildkite-stats-to-cloudwatch"
  description = "Captures Buildkite metrics and publishes them to CloudWatch"
  role = "${aws_iam_role.metrics_role.arn}"
  filename = "release/buildkite-agent-metrics-lambda-vlatest.zip"
  handler = "handler.handle"
  source_code_hash = "${base64sha256(file("release/buildkite-agent-metrics-lambda-vlatest.zip"))}"
  runtime = "go1.x"
  memory_size = 128
  timeout = 120

  environment {
    variables {
      BUILDKITE_AGENT_TOKEN  = "${var.buildkite_agent_token}"
      BUILDKITE_QUEUE = "${var.buildkite_queue}"
    }
  }
}

resource "aws_cloudwatch_event_rule" "every_minute" {
  name = "every_1_minute"
  description = "Fires every 1 minute"
  schedule_expression = "rate(1 minute)"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_invoke_lambda" {
  statement_id = "AllowExecutionFromCloudWatch"
  action = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.buildkite-agent-metrics-function.function_name}"
  principal = "events.amazonaws.com"
  source_arn = "${aws_cloudwatch_event_rule.every_minute.arn}"
}
