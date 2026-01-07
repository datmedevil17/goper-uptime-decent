# Frontend Requirements

## Overview
The frontend is a React/Next.js application that interacts with the Gopher Uptime API. It serves two main purposes: managing monitored websites for users and (optionally) displaying validator stats.

## Tech Stack
-   **Framework**: React or Next.js
-   **Styling**: Tailwind CSS (recommended) or Custom CSS
-   **State Management**: React Query (TanStack Query) or Context API
-   **Charts**: Recharts or Chart.js (for latency/uptime graphs)
-   **Auth**: JWT storage in LocalStorage/HTTP-only cookies

## Pages & Features

### 1. Landing Page (`/`)
-   **Hero Section**: Value proposition ("Decentralized Uptime Monitoring").
-   **Call to Action**: "Get Started" (Link to Login/Signup).
-   **Validator Info**: brief section on how to become a validator.

### 2. Authentication (`/login`, `/signup`)
-   **Signup Form**:
    -   Fields: Email, Password.
    -   Action: `POST /api/v1/auth/signup`.
    -   Redirect to Login or Dashboard on success.
-   **Login Form**:
    -   Fields: Email, Password.
    -   Action: `POST /api/v1/auth/login`.
    -   On success: Store JWT token and redirect to Dashboard.

### 3. Dashboard (`/dashboard`)
-   **Protected Route**: Requires valid JWT.
-   **Website List**:
    -   Fetch data from `GET /api/v1/websites`.
    -   Display grid/table of websites with:
        -   URL
        -   Current Status (Green/Red indicator)
        -   Latest Latency
        -   Last Checked timestamp
-   **Add Website Button**:
    -   Opens Modal or navigates to `/new`.
    -   Form: URL Input.
    -   Action: `POST /api/v1/website`.

### 4. Website Details (`/website/:id`)
-   **Protected Route**: Requires valid JWT.
-   **Overview**: URL, Status, Created At.
-   **Charts**:
    -   Latency over time (Line chart): Use `Ticks` from `GET /api/v1/website/status`.
    -   Uptime percentage.
-   **Actions**:
    -   **Delete Website**: `DELETE /api/v1/website` (Sends JSON body `{ "websiteId": "..." }`).

### 5. Validator Panel (Optional / Public)
-   **Balance Check**:
    -   Input: Validator ID (UUID).
    -   Action: `GET /api/v1/validator/:id/balance`.
    -   Display: Pending Payouts (USD/SOL).
-   **Request Payout**:
    -   Button to trigger `POST /api/v1/payout/:id`.

## UX/UI Notes
-   **Loading States**: Show skeletons or spinners while fetching data.
-   **Error Handling**: specific toasts/alerts for "Invalid Credentials", "Website already exists", etc.
-   **Responsive Design**: Must work on mobile and desktop.
