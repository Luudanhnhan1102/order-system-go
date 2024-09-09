# Go Assignment Project

This project implements a set of microservices and a frontend application for an e-commerce system.

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

## How to Run the Project

### Backend Services

1. Order Service:
   ```
   cd backend-order
   go run main.go
   ```
   The service will run on `http://localhost:8080`

2. Payment Service:
   ```
   cd backend-payment
   go run main.go
   ```
   The service will run on `http://localhost:8081`

### Frontend

1. Install dependencies:
   ```
   cd frontend
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

## Project Structure

- `backend-order/`: Order service implementation
- `backend-payment/`: Payment service implementation
- `frontend/`: React-based frontend application

## Technologies Used

- Backend: Go, Gin framework, PostgreSQL
- Frontend: React, TypeScript
- API Documentation: Swagger

