# 🐍 SnakeUp – Multiplayer Snake Battle Royale

SnakeUp is a **multiplayer browser-based snake battle royale** built with **Go, WebSockets, and HTML5 Canvas**.  
The project demonstrates **systems engineering** concepts: networking protocols, scaling, infra, and deployment.  

---

## 🚀 Features
- Real-time multiplayer (up to 8 players).  
- Account system + private room creation with invite links.  
- Procedurally generated maps (planned).  
- Powerups for dynamic gameplay (planned).  
- Dockerized backend + CI/CD pipeline (planned).  

---

## 🛠 Tech Stack
- **Backend**: Go + WebSockets (`gorilla/websocket` or stdlib)  
- **Frontend**: HTML5 Canvas + JavaScript + CSS  
- **Database**: PostgreSQL  
- **Infra**: Docker, GitHub Actions (CI/CD)  
- **Deployment**: Render / Fly.io / VPS  

---

## 📂 Project Structure
```
.
├── client/       # Frontend (HTML, JS, CSS)
├── server/       # Go backend (WebSocket server, DB)
├── infra/        # Docker, CI/CD configs
├── docs/         # Protocol docs, RFC-style notes
└── README.md
```

---

## ⚙️ Setup Instructions

### 1. Prerequisites
Make sure you have installed:
- [Go](https://go.dev/dl/) (>=1.20)  
- [Node.js](https://nodejs.org/) (for frontend dev, >=18 recommended)  
- [PostgreSQL](https://www.postgresql.org/)  
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized runs)  

---

### 2. Clone the Repository
```bash
git clone https://github.com/your-username/snakeup.git
cd snakeup
```

---

### 3. Backend Setup (Go Server)
1. Navigate to the server directory:
   ```bash
   cd server
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the server (default port `8080`):
   ```bash
   go run main.go
   ```

4. The server should now be live at:
   ```
   ws://localhost:8080/ws
   ```

---

### 4. Database Setup (PostgreSQL)
1. Create a new database:
   ```bash
   createdb snakeup
   ```

2. Run the migrations (example using `psql`):
   ```bash
   psql snakeup < infra/migrations/init.sql
   ```

3. Update DB connection string in `server/config.yaml` or environment variables.

---

### 5. Frontend Setup
1. Navigate to the client folder:
   ```bash
   cd client
   ```

2. Start a local server (simple example with Python):
   ```bash
   python3 -m http.server 3000
   ```

3. Open in browser:
   ```
   http://localhost:3000
   ```

---

### 6. Run with Docker (Optional)
To run the full stack in Docker:
```bash
docker-compose up --build
```
This will spin up:
- Backend (Go server)  
- PostgreSQL database  
- Frontend (served as static files)  

---

## 🧪 Development Workflow
- Write Go code → hot reload with `air` (if installed).  
- Frontend testing → live reload with `vite` or simple HTTP server.  
- Commit + push → GitHub Actions runs tests and builds.  

---

## 📜 Roadmap
- [x] Local snake prototype  
- [ ] Multiplayer WebSocket server  
- [ ] Account system + rooms  
- [ ] Procedural maps & powerups  
- [ ] Docker + CI/CD pipeline  
- [ ] Load testing with bots  

---

## 🤝 Contributing
PRs are welcome! Please open an issue first to discuss new features or bug fixes.  
