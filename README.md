# AMS Backend

Backend API for the **Academic Management System (AMS)**.

The backend provides authentication, role-based authorization, faculty and student management, course management, faculty-course assignments, course registration, academic records, grading, and grade notification emails.

## Features

* User authentication using JWT
* HTTP-only authentication cookies
* Role-based authorization

  * Admin
  * Faculty
  * Student
* Admin management of:

  * Faculty
  * Students
  * Courses
  * Faculty-course assignments
  * Course registration windows
* Faculty features:

  * Faculty details
  * Current courses
  * Teaching history
  * Assigned courses and terms
  * Course student lists
  * Grade roster
  * Grade submission
* Student features:

  * Student details
  * Academic records
  * Available courses
  * Course registration
  * Registration history
* Email notification when grades are posted or updated
* CORS support for the frontend

## Tech Stack

| Technology | Purpose                               |
| ---------- | ------------------------------------- |
| Go         | Backend programming language          |
| Fiber v2   | HTTP web framework                    |
| PostgreSQL | Relational database                   |
| pgx/v5     | PostgreSQL driver and connection pool |
| JWT        | Authentication                        |
| bcrypt     | Password hashing/verification         |
| Resend     | Grade notification emails             |
| godotenv   | Loading environment variables         |

The current `go.mod` specifies Go `1.26.5` and includes Fiber, JWT, pgx, godotenv, SendGrid, and crypto dependencies. The current email implementation uses the Resend HTTP API.

## Project Structure

```text
AMS-backend/
│
├── db/
│   └── db.go
│
├── handlers/
│   ├── assignment.go
│   ├── auth.go
│   ├── course.go
│   ├── facultyDashboard.go
│   ├── facultyStudents.go
│   └── studentRegisterCourse.go
│
├── middleware/
│   └── auth.go
│
├── models/
│   ├── assignment.go
│   ├── course.go
│   └── facultyTeaching.go
│
├── utils/
│   ├── email.go
│   └── jwt.go
│
├── main.go
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

Install the following before running the backend:

* Go
* PostgreSQL
* Git

Verify Go:

```bash
go version
```

Verify PostgreSQL:

```bash
psql --version
```

## Clone the Repository

```bash
git clone https://github.com/Srinivasan3009/AMS-backend.git
cd AMS-backend
```

## Environment Variables

Create a `.env` file in the project root.

```env
DATABASE_URL=postgresql://username:password@localhost:5432/ams
JWT_SECRET=your-secure-jwt-secret
FRONTEND_URL=http://localhost:5173
PORT=8080
RESEND_API_KEY=your-resend-api-key
```

## Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE ams;
```

The backend expects the database schema required by the application, including tables such as:

## Database Tables

The AMS backend uses PostgreSQL. The main database tables are:

| Table | Purpose |
|---|---|
| `users` | Stores login credentials, user identity, and roles |
| `students` | Stores student-specific information |
| `faculty` | Stores faculty-specific information |
| `courses` | Stores course information |
| `assignments` | Stores faculty-course teaching assignments |
| `course_registration_windows` | Controls when students can register for courses |
| `course_registrations` | Stores students' course registration details |
| `academic_records` | Stores students' academic results and grades |
| `faculty_teaching` | Stores faculty teaching history/details |

The backend connects to PostgreSQL using `pgxpool` and expects `DATABASE_URL` to be configured.

> **Assumption:** The required database schema/data is managed separately from this repository. If a database schema SQL file is provided by the project, execute it before starting the backend.

## Install Dependencies

From the project directory:

```bash
go mod download
```

Or:

```bash
go mod tidy
```

## Run the Backend

```bash
go run main.go
```

The server runs on:

```text
http://localhost:8080
```

unless another port is supplied through the `PORT` environment variable.

You should see a database connection message followed by the Fiber server starting.


## Authentication

Authentication uses JWT tokens.

After successful login:

1. The backend validates the supplied credentials.
2. The password is checked against the stored bcrypt hash.
3. A JWT containing the user ID and role is generated.
4. The JWT is stored in an HTTP-only `token` cookie.
5. Protected routes validate the cookie before processing the request.

JWT tokens currently expire after **24 hours**.

Protected endpoints return:

```json
{
  "error": "unauthorized"
}
```

when no valid authentication cookie is supplied.

Role-restricted endpoints return:

```json
{
  "error": "forbidden"
}
```

when the authenticated user's role does not match the required role.

## Course Management

Courses can be created and updated by administrators.

The backend calculates TCP automatically:

```text
TCP = Lecture Hours + Tutorial Hours + Practical Hours
```

TCP is calculated server-side instead of trusting the value sent by the client.

Courses can also be filtered using query parameters such as:

```text
department
semester
batch
category
```

Example:

```text
GET /api/admin/courses?department=CSE&semester=5
```

## Grade Notifications

When a faculty member submits grades, the system can send an email notification to the student.

The email service uses the **Resend HTTP API** and requires:

```env
RESEND_API_KEY=your-resend-api-key
```

The current implementation sends from:

```text
Anna University Portal <onboarding@resend.dev>
```

and uses Resend's API rather than SMTP.

If email notifications are required, make sure a valid Resend API key is configured.

## Grade Notification Delivery Logging

When a faculty member submits or updates a student's grade, the system attempts to notify the student by email.

The outcome of each notification attempt is recorded in the faculty member's email log:

* **Success** — records that the notification was successfully sent.
* **Failure** — records that the notification attempt failed.
* The log provides a permanent record of the notification outcome for each student.
* Faculty members can review the notification status from within the portal without needing to check an external email service.

This provides visibility into whether each student was actually notified and allows faculty members to identify unsuccessful notification attempts.

### Notification Flow

```text
Faculty submits/updates grade
            │
            ▼
      Backend processes grade
            │
            ▼
      Email notification
            │
       ┌────┴────┐
       │         │
    Success    Failure
       │         │
       └────┬────┘
            ▼
   Record notification
       outcome in log
            │
            ▼
      Faculty can view
       delivery status
```

The notification log is maintained by the backend so that delivery outcomes remain available as part of the application's academic/notification records.


## CORS Configuration

The backend reads the frontend URL from:

```env
FRONTEND_URL=http://localhost:5173
```

The API allows credentials so that the authentication cookie can be used by the frontend.

When using a local React/Vite frontend, configure the appropriate frontend origin:

```env
FRONTEND_URL=http://localhost:5173
```

## Testing the API

The APIs can be tested using:

* Postman
* Thunder Client
* Insomnia
* Frontend application

Example login request:

```http
POST http://localhost:8080/api/login
Content-Type: application/json
```

Example request body:

```json
{
  "identifier": "your-username-or-register-number",
  "password": "your-password"
}
```

After successful authentication, the server sets the `token` cookie.

You can then access:

```http
GET http://localhost:8080/api/me
```

using the same authenticated session.

## Running with a Frontend

The backend is designed to work with a separate frontend application.

Typical local setup:

```text
Frontend
   │
   │ HTTP requests
   ▼
Go Fiber Backend
   │
   ├── JWT Authentication
   ├── Role Authorization
   │
   ▼
PostgreSQL Database
```

For local development, make sure the frontend URL matches the value configured in `FRONTEND_URL`.

## Important Assumptions

1. **PostgreSQL is available** and the database schema required by the application has already been created.
2. The `DATABASE_URL` environment variable points to a valid PostgreSQL database.
3. User passwords stored in the database are bcrypt hashes.
4. User roles are represented by the expected values:

   * `admin`
   * `faculty`
   * `student`
5. `JWT_SECRET` is kept private and is not committed to source control.
6. The frontend sends requests with credentials enabled so the HTTP-only authentication cookie can be used.
7. The frontend origin must be configured correctly through `FRONTEND_URL`.
8. Grade email notifications require a valid `RESEND_API_KEY`.
9. The backend currently uses a 24-hour JWT expiration period.
10. The database schema and initial users/data are assumed to be provisioned separately from the backend source code.
11. The application is intended to be used over HTTPS in deployed environments because authentication cookies are configured as secure cookies.


## License

This project does not currently specify a license in the repository.

## Author

**Srinivasan S**

GitHub:
https://github.com/Srinivasan3009
