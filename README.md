# 🚀 AsyncProd

**AsyncProd** is a backend service built in Go for asynchronously processing product-related operations like uploading, compressing, and storing product images. It integrates with RabbitMQ for message queueing and AWS S3 for storing processed images. The service also includes APIs for managing product data and health check endpoints for system monitoring.

---
<div align="center">
  <img src="https://github.com/sameerihs/AsyncProd/blob/main/async-prod.png" alt="AsyncProd Workflow" width="600">
</div>

## Features

- **Product Management**:
  - Create, Read, Update, and Query products.
  - Includes validation for user input.

- **Asynchronous Image Processing**:
  - Compresses product images to optimize size.
  - Uploads compressed images to AWS S3.
  - Utilizes RabbitMQ for message queueing.

- **Scalable Architecture**:
  - Database: PostgreSQL for relational data.
  - Redis for caching.
  - RabbitMQ for asynchronous task processing.
  - AWS S3 for file storage.

- **Health Monitoring**:
  - API endpoints to check the health of PostgreSQL and Redis.

---

## Prerequisites

Ensure the following dependencies are installed:

- **Go**: v1.20+
- **RabbitMQ**: v3.9+
- **PostgreSQL**: v14+
- **Redis**: v6+
- **AWS S3**: Configured with access/secret keys and a bucket.

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-repo/asyncprod.git
cd asyncprod
```

### 2. Set Up Environment Variables
Create a .env file in the root directory and populate it with the following configuration:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=asyncprod_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_IMAGE_QUEUE=image_processing_queue

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key
AWS_BUCKET_NAME=your_bucket_name
```
### 3. Install Dependencies
```bash
go mod tidy
```
### 4. Initialize and Run the Application
```bash
go run main.go
```

---

## API Endpoints

### Products
#### 1. Create a Product


```bash
POST /api/v1/products

{
  "user_id": 1,
  "product_name": "Sample Product",
  "product_description": "This is a sample product",
  "product_price": 100.50,
  "product_images": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ]
}

```

#### 2. Get a Product by ID
```bash 
GET /api/v1/products/:id
```
#### 3. Get All Products by User ID

``` bash
GET /api/v1/products?user_id=1&min_price=0&max_price=500&product_name=Sample
```

#### 4. Update a Product
``` bash
PUT /api/v1/products
```

---
# 🏗️ Architecture Overview

## Go Modules

- **`models`**: Handles all database interactions, including creating, reading, updating, and deleting records.
- **`services`**: Manages the core application logic, such as image processing and communication with RabbitMQ.
- **`config`**: Centralized configuration for external dependencies, including:
  - Database (PostgreSQL)
  - Redis (Caching)
  - RabbitMQ (Message Queue)
  - AWS S3 (Storage)

---

## Message Queue

- **Publish Messages**: When a new product is created, the service publishes a message to a RabbitMQ queue.
- **Worker Service**: A separate service listens to the queue and processes messages asynchronously, ensuring smooth task handling and decoupling.

---

## 🖼️ Image Processing Workflow

### 1. Publish Message
- When a product is created, the **image URLs** are sent to a **RabbitMQ queue** for processing.

### 2. Message Consumption
- A **RabbitMQ consumer** fetches the message from the queue and downloads the images using the provided URLs.

### 3. Image Compression
- Images are resized to a **maximum resolution of 800x600**.
- JPEG quality is set to **75%** to ensure a balance between quality and file size.

### 4. Upload to AWS S3
- The compressed images are uploaded to an **S3 bucket**.
- **Public URLs** for the uploaded images are generated and stored for later use.

### 5. Update Product
- The product record in the database is updated with the new **compressed image URLs**.

