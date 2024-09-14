provider "aws" {
  region = "us-east-1" 
}

# Use the default VPC
data "aws_vpc" "default" {
  default = true
}

# Use default subnets
data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "cursor-test-cluster"
}

# ECR Repositories
resource "aws_ecr_repository" "backend_order" {
  name = "cursor-test-backend-order"
  force_delete = true
}

resource "aws_ecr_repository" "backend_payment" {
  name = "cursor-test-backend-payment"
  force_delete = true
}

# Default Security Group (if not already defined)
data "aws_security_group" "default" {
  name   = "default"
  vpc_id = data.aws_vpc.default.id
}

# ECS Task Definitions
resource "aws_ecs_task_definition" "backend_order" {
  family                   = "cursor-test-backend-order"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([{
    name  = "cursor-test-backend-order"
    image = "${aws_ecr_repository.backend_order.repository_url}:latest"
    portMappings = [{
      containerPort = 8080
      hostPort      = 8080
    }]
    environment = [
      {
        name  = "MONGODB_URI"
        value = "mongodb://${aws_docdb_cluster.default.master_username}:${aws_docdb_cluster.default.master_password}@${aws_docdb_cluster.default.endpoint}:27017/?tls=false&retryWrites=false"
      },
      {
        name  = "MONGODB_DATABASE"
        value = "backend-order"
      },
      {
        name  = "PORT"
        value = "8080"
      },
      {
        name  = "API_URL"
        value = "https://cursor-experiment-api-order.sotalabs.io"
      },
      {
        name  = "API_PAYMENT_URL"
        value = "http://backend-payment.cursor-test.internal:8081"
      },
      {
        name  = "MAILTRAP_API_TOKEN"
        value = "2c0514badc6d435a9e09bb4b2584a048"
      },
      {
        name  = "API_SECRET_KEY"
        value = "secret-key-for-backend-call"
      }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = "/ecs/cursor-test-backend-order"
        awslogs-region        = "us-east-1"
        awslogs-stream-prefix = "ecs"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "backend_payment" {
  family                   = "cursor-test-backend-payment"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([{
    name  = "cursor-test-backend-payment"
    image = "${aws_ecr_repository.backend_payment.repository_url}:latest"
    portMappings = [{
      containerPort = 8081
      hostPort      = 8081
    }]
    environment = [
      {
        name  = "MONGODB_URI"
        value = "mongodb://${aws_docdb_cluster.default.master_username}:${aws_docdb_cluster.default.master_password}@${aws_docdb_cluster.default.endpoint}:27017/?tls=false&retryWrites=false"
      },
      {
        name  = "MONGODB_DATABASE"
        value = "backend-payment"
      },
      {
        name  = "PORT"
        value = "8081"
      },
      {
        name  = "API_URL"
        value = "https://cursor-experiment-api-payment.sotalabs.io"
      },
      {
        name  = "API_ORDER_URL"
        value = "http://backend-order.cursor-test.internal:8080"
      },
      {
        name  = "API_SECRET_KEY"
        value = "secret-key-for-backend-call"
      }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = "/ecs/cursor-test-backend-payment"
        awslogs-region        = "us-east-1"
        awslogs-stream-prefix = "ecs"
      }
    }
  }])
}

# ECS Services
resource "aws_ecs_service" "backend_order" {
  name            = "cursor-test-backend-order"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.backend_order.arn
  launch_type     = "FARGATE"
  desired_count   = 1

  network_configuration {
    assign_public_ip = true
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.ecs_tasks.id, data.aws_security_group.default.id]
  }

  force_new_deployment = true

  load_balancer {
    target_group_arn = aws_lb_target_group.backend_order.arn
    container_name   = "cursor-test-backend-order"
    container_port   = 8080
  }

  service_registries {
    registry_arn = aws_service_discovery_service.backend_order.arn
  }
}

resource "aws_ecs_service" "backend_payment" {
  name            = "cursor-test-backend-payment"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.backend_payment.arn
  launch_type     = "FARGATE"
  desired_count   = 1

  network_configuration {
    assign_public_ip = true
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.ecs_tasks.id, data.aws_security_group.default.id]
  }

  force_new_deployment = true

  load_balancer {
    target_group_arn = aws_lb_target_group.backend_payment.arn
    container_name   = "cursor-test-backend-payment"
    container_port   = 8081
  }

  service_registries {
    registry_arn = aws_service_discovery_service.backend_payment.arn
  }
}

# Security Group for ECS Tasks
resource "aws_security_group" "ecs_tasks" {
  name        = "cursor-test-ecs-tasks-sg"
  description = "Allow inbound traffic for ECS tasks"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8081
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "cursor-test-ecs-tasks-sg"
  }
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "cursor-test-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

# Attach necessary policies to the IAM role
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_ecr_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy_logs" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "backend_order" {
  name              = "/ecs/cursor-test-backend-order"
  retention_in_days = 30
}

# Add a CloudWatch Log Group for backend-payment
resource "aws_cloudwatch_log_group" "backend_payment" {
  name              = "/ecs/cursor-test-backend-payment"
  retention_in_days = 30
}

# DocumentDB Subnet Group
resource "aws_docdb_subnet_group" "default" {
  name       = "cursor-test-docdb-subnet-group"
  subnet_ids = data.aws_subnets.default.ids

  tags = {
    Name = "Cursor Test DocumentDB subnet group"
  }
}

# DocumentDB Cluster Parameter Group
resource "aws_docdb_cluster_parameter_group" "no_tls" {
  family = "docdb5.0"
  name   = "cursor-test-docdb-no-tls"

  parameter {
    name  = "tls"
    value = "disabled"
  }
}

# DocumentDB Cluster
resource "aws_docdb_cluster" "default" {
  cluster_identifier      = "cursor-test-docdb-cluster"
  engine                  = "docdb"
  master_username         = "docdbadmin"
  master_password         = "YourStrongPasswordHere"  # Consider using a secret manager for production
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  skip_final_snapshot     = true
  db_subnet_group_name    = aws_docdb_subnet_group.default.name
  vpc_security_group_ids  = [aws_security_group.docdb.id]
  db_cluster_parameter_group_name = aws_docdb_cluster_parameter_group.no_tls.name
}

# DocumentDB Cluster Instance
resource "aws_docdb_cluster_instance" "cluster_instances" {
  count              = 1
  identifier         = "cursor-test-docdb-instance-${count.index}"
  cluster_identifier = aws_docdb_cluster.default.id
  instance_class     = "db.t3.medium"
}

# Outputs
output "backend_order_url" {
  value = "https://cursor-experiment-api-order.sotalabs.io"
}

output "backend_payment_url" {
  value = "https://cursor-experiment-api-payment.sotalabs.io"
}

output "docdb_endpoint" {
  value = aws_docdb_cluster.default.endpoint
}

output "backend_order_alb_dns" {
  value = aws_lb.backend_order.dns_name
}

output "backend_payment_alb_dns" {
  value = aws_lb.backend_payment.dns_name
}

# Security Group for DocumentDB
resource "aws_security_group" "docdb" {
  name        = "cursor-test-docdb-sg"
  description = "Allow inbound traffic to DocumentDB from anywhere"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "Allow traffic from anywhere"
    from_port   = 27017
    to_port     = 27017
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "cursor-test-docdb-sg"
  }
}

# Add a rule to allow traffic from the default security group to DocumentDB
resource "aws_security_group_rule" "docdb_from_default_sg" {
  type                     = "ingress"
  from_port                = 27017
  to_port                  = 27017
  protocol                 = "tcp"
  security_group_id        = aws_security_group.docdb.id
  source_security_group_id = data.aws_security_group.default.id
}

# S3 bucket for frontend hosting
resource "aws_s3_bucket" "frontend" {
  bucket = "cursor-test-frontend"
  force_destroy = true
}

resource "aws_s3_bucket_website_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "index.html"
  }
}

resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.frontend.arn}/*"
      },
    ]
  })
}

# CloudFront distribution for frontend
resource "aws_cloudfront_distribution" "frontend" {
  origin {
    domain_name = aws_s3_bucket_website_configuration.frontend.website_endpoint
    origin_id   = "S3-${aws_s3_bucket.frontend.bucket}"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-${aws_s3_bucket.frontend.bucket}"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  aliases = ["cursor-experiment.sotalabs.io"]

  viewer_certificate {
    acm_certificate_arn = "arn:aws:acm:us-east-1:975050238074:certificate/fd5825a5-26b3-449d-9d56-26f4ec14bba0"
    ssl_support_method  = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}

# Update the frontend_url output
output "frontend_url" {
  value = "https://cursor-experiment.sotalabs.io"
}

output "cloudfront_domain_name" {
  value = aws_cloudfront_distribution.frontend.domain_name
}

data "aws_region" "current" {}

# Application Load Balancer
resource "aws_lb" "backend_order" {
  name               = "cursor-test-alb-order"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = data.aws_subnets.default.ids

  enable_deletion_protection = false
}

resource "aws_lb" "backend_payment" {
  name               = "cursor-test-alb-payment"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = data.aws_subnets.default.ids

  enable_deletion_protection = false
}

# Security Group for ALB
resource "aws_security_group" "alb" {
  name        = "cursor-test-alb-sg"
  description = "Allow inbound traffic to ALB"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Target Groups
resource "aws_lb_target_group" "backend_order" {
  name        = "cursor-test-tg-backend-order"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    path                = "/health"
    healthy_threshold   = 2
    unhealthy_threshold = 10
    timeout             = 60
    interval            = 300
    matcher             = "200"
  }
}

resource "aws_lb_target_group" "backend_payment" {
  name        = "cursor-test-tg-backend-payment"
  port        = 8081
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    path                = "/health"
    healthy_threshold   = 2
    unhealthy_threshold = 10
    timeout             = 60
    interval            = 300
    matcher             = "200"
  }
}

# Listeners
resource "aws_lb_listener" "backend_order" {
  load_balancer_arn = aws_lb.backend_order.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "backend_payment" {
  load_balancer_arn = aws_lb.backend_payment.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# HTTPS Listener for backend_order
resource "aws_lb_listener" "backend_order_https" {
  load_balancer_arn = aws_lb.backend_order.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "arn:aws:acm:us-east-1:975050238074:certificate/fd5825a5-26b3-449d-9d56-26f4ec14bba0"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend_order.arn
  }
}

# HTTPS Listener for backend_payment
resource "aws_lb_listener" "backend_payment_https" {
  load_balancer_arn = aws_lb.backend_payment.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "arn:aws:acm:us-east-1:975050238074:certificate/fd5825a5-26b3-449d-9d56-26f4ec14bba0"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend_payment.arn
  }
}

# Service Discovery Namespace
resource "aws_service_discovery_private_dns_namespace" "main" {
  name        = "cursor-test.internal"
  description = "Private DNS namespace for cursor-test services"
  vpc         = data.aws_vpc.default.id
}

# Service Discovery Service for backend-order
resource "aws_service_discovery_service" "backend_order" {
  name = "backend-order"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

# Service Discovery Service for backend-payment
resource "aws_service_discovery_service" "backend_payment" {
  name = "backend-payment"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}
