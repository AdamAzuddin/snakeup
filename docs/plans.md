# Project Plan: SnakeUp (Multiplayer Snake Battle Royal)
BMAD Format – Background, Motivation, Aim, Details

---

## 🟢 Background
I want to build this project to learn hard technical concepts. CRUD apps no longer challenge me, and with AI becoming more prevalent, I need to level up my engineering skills so I can use AI as a tool rather than a crutch.  

My skills today:  
- **Go**: Beginner (CLI tools, CRUD server, Postgres integration).  
- **Web Development**: Above intermediate (20 years of experience building websites, currently using these skills in internship).  
- **System Design**: Beginner, but motivated to stand out here.  
- **Infra/DevOps**: Some Docker experience from a multilingual chess project (Unity WebGL + Python).  

I can realistically commit **10–20 hours/week over 5 months**, with flexibility around exams.

---

## 🟠 Motivation
I want this project to prove that I am:  
1. A reliable person who can **figure things out and ship real projects**.  
2. Capable of tackling **systems complexity** (protocols, scaling, infra).  
3. More interested in **serious system engineering** than just designing pretty UIs or static maps.  

I’m most excited about building the **networking protocols** and **scaling the system**, followed by infra. Game maps will be **procedurally generated** so I don’t spend time manually designing levels.

Impression I want people to have:  
> “This is a serious systems engineering project that happens to be a fun multiplayer game.”

---

## 🎯 Aim
**Minimum Success**  
- A functional **multiplayer snake game** (up to 8 players).  
- Users can create accounts, create private rooms, and invite friends via link.  

**Stretch Success**  
- **Procedurally generated maps** with multiple levels that increase difficulty.  
- At least **2–3 unique powerups** that make gameplay fun and dynamic.  

**Measurement of Success**  
- Technical completion (the system works, scales, and is deployed).  
- Real user validation (friends can actually play it online).  

---

## ⚙️ Details

### Tech Stack
- **Backend**: Go + WebSocket (`gorilla/websocket` or stdlib).  
- **Frontend**: Minimal stack — HTML5 Canvas + JS + CSS.  
- **Database**: PostgreSQL (accounts, high scores, friendships).  
- **Infra**: Docker for containerization. CI/CD with GitHub Actions after MVP.  
- **Deployment**: Cloud hosting (Render, Fly.io, or VPS).  

---

### Project Roadmap (Phases)

#### Phase 1 – Core Gameplay (2–3 weeks)
- Build basic snake game in browser (HTML5 Canvas).  
- Local only: move snake, eat food, die on collision.  

#### Phase 2 – Multiplayer Prototype (3–4 weeks)
- Implement Go WebSocket server.  
- Sync state: player positions, food, collisions.  
- Support 2–4 players in one arena (janky is okay).  

#### Phase 3 – Battle Royale Core (4 weeks)
- Expand arena to handle up to 8 players.  
- Add **room creation**: a player generates a link, friends join.  
- Simple account system (register/login with Postgres).  
- Leaderboard (highest scores).  

#### Phase 4 – Procedural Maps + Powerups (5–6 weeks)
- Procedurally generate maps with obstacles and multiple levels.  
- Add 2–3 powerups (e.g., teleport, clone, enemy shrink).  
- Improve collision and fairness handling.  

#### Phase 5 – Infra & Scaling (3–4 weeks)
- Dockerize backend + database.  
- Deploy live (Render/Fly.io).  
- Add CI/CD (GitHub Actions for build/test/deploy).  
- Load testing with bots (simulate 8 players).  

#### Phase 6 – Polish & Extras (time permitting)
- Add friend system (optional).  
- Skins/colors for snakes.  
- Room chat.  
- Promotion online (GitHub, LinkedIn, maybe product demo).  

---

### Deliverables
- **Codebase**: Organized repo with `client/`, `server/`, `docs/`, `infra/`.  
- **Docs**: RFC-style protocol description (how the WebSocket protocol works).  
- **Deployment**: Live version friends can play.  
- **Portfolio Value**: Serious system engineering + real playable game.  

---

## ✅ Summary
Snake Royale is a multiplayer browser-based game built with Go + WebSocket and deployed with Docker. The focus is on **systems engineering** (protocols, scaling, infra) rather than just gameplay. Success is defined by both **technical completion** and **real players being able to join and have fun online**.  
