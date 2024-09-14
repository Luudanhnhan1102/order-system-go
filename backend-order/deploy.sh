#!/bin/bash

# Set variables
SERVICE_NAME="backend-order"  # Change this to "backend-payment" for the payment service
CLUSTER_NAME="cursor-test-cluster"
ECR_REPO="975050238074.dkr.ecr.us-east-1.amazonaws.com/cursor-test-${SERVICE_NAME}"

# Authenticate Docker to Amazon ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 975050238074.dkr.ecr.us-east-1.amazonaws.com

# Build the Docker image
docker build -t ${SERVICE_NAME} .

# Tag the image
docker tag ${SERVICE_NAME}:latest ${ECR_REPO}:latest

# Push the image to ECR
docker push ${ECR_REPO}:latest

echo "Image pushed to ECR successfully!"

# Get the current task definition
TASK_DEF_ARN=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services cursor-test-${SERVICE_NAME} --query 'services[0].taskDefinition' --output text)
echo "Current task definition ARN: ${TASK_DEF_ARN}"

# Retrieve the task definition
aws ecs describe-task-definition --task-definition ${TASK_DEF_ARN} --query 'taskDefinition' > task-definition.json

# Update the image in the task definition
jq ".containerDefinitions[0].image = \"${ECR_REPO}:latest\"" task-definition.json > updated-task-definition.json

# Remove unnecessary fields from the task definition
jq 'del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy)' updated-task-definition.json > cleaned-task-definition.json

# Register the new task definition
NEW_TASK_DEF_ARN=$(aws ecs register-task-definition --cli-input-json file://cleaned-task-definition.json --query 'taskDefinition.taskDefinitionArn' --output text)
echo "New task definition ARN: ${NEW_TASK_DEF_ARN}"

# Update the service to use the new task definition
aws ecs update-service --cluster ${CLUSTER_NAME} --service cursor-test-${SERVICE_NAME} --task-definition ${NEW_TASK_DEF_ARN} | jq

echo "ECS service updated successfully!"

# Clean up temporary files
rm task-definition.json updated-task-definition.json cleaned-task-definition.json

echo "Deployment completed!"