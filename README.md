# Go Assignment Project

This project implements a microservices-based e-commerce system with a React frontend.

## Components

### Backend Services

1. **Order Service** (backend-order)
   - Manages customer orders
   - Handles product catalog
   - User authentication and authorization

2. **Payment Service** (backend-payment)
   - Processes payments for orders

### Frontend

- React-based web application
- Provides user interface for interacting with the backend services

## Project Structure

- `backend-order/`: Order service implementation
- `backend-payment/`: Payment service implementation
- `frontend/`: React-based frontend application
- `infra/`: Terraform configuration for AWS infrastructure

## Technologies Used

- Backend: Go, Gin framework, MongoDB
- Frontend: React, TypeScript
- Infrastructure: AWS (ECS, ECR, ALB, DocumentDB, S3, CloudFront)
- API Documentation: Swagger
- IaC: Terraform

## How to Run the Project Locally

### Backend Services

1. Order Service:
   ```
   cd backend-order
   cp .env.example .env
   go run main.go
   ```
   The service will run on `http://localhost:8080`

2. Payment Service:
   ```
   cd backend-payment
   cp .env.example .env
   go run main.go
   ```
   The service will run on `http://localhost:8081`

### Frontend

1. Install dependencies:
   ```
   cd frontend
   cp .env.local .env
   npm install
   ```

2. Start the development server:
   ```
   npm start
   ```
   The frontend will be available at `http://localhost:3000`

## API Documentation

- Order Service Swagger UI: `http://localhost:8080/swagger/index.html`
- Payment Service Swagger UI: `http://localhost:8081/swagger/index.html`

## ProductionDeployment

The project is deployed on AWS using Terraform. The infrastructure includes:

- ECS Fargate for running backend services
- ECR for Docker image storage
- Application Load Balancers for backend services
- DocumentDB for database
- S3 and CloudFront for frontend hosting

To deploy the infrastructure:

1. Navigate to the `infra` directory
2. Initialize Terraform:
   ```
   terraform init
   ```
3. Apply the Terraform configuration:
   ```
   terraform apply
   ```

### Deploy Backend Services

1. Navigate to the `backend-order` directory
2. Run the deployment script
   ```
   cd backend-order
   ./deploy.sh
   ```

### Deploy Frontend

1. Navigate to the `frontend` directory
2. Run the deployment script
   ```
   cd frontend
   npm run deploy:prod
   ```

## Environment Variables

Both backend services use environment variables for configuration. Ensure these are set in your local `.env` files and in the AWS ECS task definitions.

Key variables include:
- `MONGODB_URI`
- `MONGODB_DATABASE`
- `PORT`
- `API_URL`
- `API_PAYMENT_URL` (for Order Service)
- `API_ORDER_URL` (for Payment Service)
- `MAILTRAP_API_TOKEN` (for Order Service)

## Service Discovery

The backend services use AWS Cloud Map for service discovery, allowing them to communicate using internal DNS names within the VPC.

## Security

- HTTPS is enforced for all public endpoints
- MongoDB connections use TLS
- IAM roles are used for ECS task execution

## Monitoring and Logging

CloudWatch is used for monitoring and logging of ECS tasks and other AWS resources.
