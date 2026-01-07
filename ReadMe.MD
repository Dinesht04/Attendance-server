# Live Attendance System (Backend)

A real-time attendance management system built with **Go (Golang)**, **WebSockets**, and **MongoDB**. This project was developed as part of a backend engineering assignment to handle live student attendance updates and class management via a persistent connection. ([Assignment's Notion Doc Link](https://brindle-goal-102.notion.site/Backend-WebSocket-Live-Attendance-System-2c646b36b2e980b09b42d7c0240a8170))

## Features

* **Real-time Communication**: Utilizes WebSockets for live attendance updates and broadcasting.
* **Class Management**: Create and manage virtual classrooms.
* **Student Management**: Track student profiles and their enrollment status.
* **Attendance Tracking**: Automated/Real-time attendance logging.
* **JWT Authentication**: Secure endpoints for admins and students (as structured in `auth.go`).
* **Database Persistence**: Integration with MongoDB for storing attendance logs, student records, and class data.

## Project Structure

```text
.
├── data/
│   ├── data.go         
│   └── model.go
├── server/
│   ├── attendance.go  
│   ├── auth.go        
│   ├── class.go        
│   ├── hub.go          
│   ├── server.go       
│   ├── student.go      
│   └── websocket.go    
├── util/
│   └── error.go        
├── go.mod              
├── go.sum              
├── main.go            
├── ReadMe.MD          

```

## 🛠️ Installation & Setup

### Prerequisites

* **Go**: Version 1.18 or higher installed.
* **MongoDB**: A running MongoDB instance (local or Atlas).

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd <project-folder>

```

### 2. Configure Environment

Ensure your MongoDB URI is correctly configured in `data/data.go` or via environment variables (depending on your implementation).

### 3. Install Dependencies

```bash
go mod tidy

```

### 4. Run the Server

```bash
go run main.go

```

The server will start on `http://localhost:8080`.

## Assignment Requirements Implemented

* [x] **Connection Management**: Multi-client support via `hub.go`.
* [x] **Data Persistence**: All records saved to MongoDB.
* [x] **Concurrency**: Efficient use of Goroutines for handling WebSocket messages.
* [x] **Error Handling**: Centralized error responses using the `util` package.

## 🛠 Tech Stack

* **Language**: Go (Golang)
* **Database**: MongoDB
* **Protocol**: WebSockets & REST
* **Auth**: JWT (JSON Web Tokens)

---

*Developed as a Backend Assignment.*