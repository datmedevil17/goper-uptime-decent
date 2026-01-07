# API Routes Documentation

Base URL: `http://localhost:8080`

## Authentication

### Signup
Create a new user account.
-   **URL**: `/api/v1/auth/signup`
-   **Method**: `POST`
-   **Auth**: Public
-   **Body**:
    ```json
    {
      "email": "user@example.com",
      "password": "securepassword123"
    }
    ```
-   **Response** (`201 Created`):
    ```json
    {
      "token": "eyJhbG...",
      "user": {
        "id": "uuid...",
        "email": "user@example.com"
      }
    }
    ```

### Login
Authenticate an existing user.
-   **URL**: `/api/v1/auth/login`
-   **Method**: `POST`
-   **Auth**: Public
-   **Body**:
    ```json
    {
      "email": "user@example.com",
      "password": "securepassword123"
    }
    ```
-   **Response** (`200 OK`):
    ```json
    {
      "token": "eyJhbG...",
      "user": {
        "id": "uuid...",
        "email": "user@example.com"
      }
    }
    ```

## Website Management
**Requires Authentication Header**: `Authorization: Bearer <token>`

### Create Website
Register a new website for monitoring.
-   **URL**: `/api/v1/website`
-   **Method**: `POST`
-   **Body**:
    ```json
    {
      "url": "https://google.com"
    }
    ```
-   **Response** (`201 Created`):
    ```json
    {
      "id": "uuid...",
      "url": "https://google.com"
    }
    ```

### List Websites
Get all active websites for the authenticated user, including recent stats.
-   **URL**: `/api/v1/websites`
-   **Method**: `GET`
-   **Response** (`200 OK`):
    ```json
    {
      "count": 1,
      "websites": [
        {
          "ID": "...",
          "URL": "https://google.com",
          "Ticks": [ ... ]
        }
      ]
    }
    ```

### Get Website Status
Get detailed status and history for a specific website.
-   **URL**: `/api/v1/website/status`
-   **Method**: `GET`
-   **Query Params**: `?websiteId=<uuid>`
-   **Response** (`200 OK`):
    ```json
    {
      "ID": "...",
      "URL": "...",
      "Ticks": [
        {
          "Status": "Good",
          "Latency": 120,
          "CreatedAt": "..."
        }
      ]
    }
    ```

### Delete Website
Stop monitoring a website (soft delete).
-   **URL**: `/api/v1/website`
-   **Method**: `DELETE`
-   **Body**:
    ```json
    {
      "websiteId": "uuid..."
    }
    ```
-   **Response** (`200 OK`):
    ```json
    {
      "message": "Website deleted successfully"
    }
    ```

## Validator & Payouts

### Request Payout
Queue a payout for accumulated validator rewards.
-   **URL**: `/api/v1/payout/:validatorId`
-   **Method**: `POST`
-   **Auth**: Public (logic checks validator balance)
-   **Response** (`200 OK`):
    ```json
    {
      "status": "queued",
      "amount": 500.00
    }
    ```

### Get Validator Balance
Check current pending balance.
-   **URL**: `/api/v1/validator/:validatorId/balance`
-   **Method**: `GET`
-   **Auth**: Public
-   **Response** (`200 OK`):
    ```json
    {
      "validator_id": "...",
      "pending_payouts": 5000000000,
      "pending_payouts_sol": 5.0
    }
    ```

## System

### Health Check
-   **URL**: `/health`
-   **Method**: `GET`
-   **Response** (`200 OK`):
    ```json
    {
      "status": "ok",
      "service": "uptime-monitor-api"
    }
    ```
