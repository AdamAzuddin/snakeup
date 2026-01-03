// SnakeUp - Simple Snake game using HTML Canvas
// The game is modular and easy to extend.

(() => {
  // -----------------------------
  // Constants and configuration
  // -----------------------------
  const canvas = document.getElementById("gameCanvas");
  const ctx = canvas.getContext("2d");

  const WORLD_COLS = 50;
  const WORLD_ROWS = 50;

  const SNAKE_COLOR = "#1976d2"; // blue (fallback)
  const APPLE_COLOR = "#d32f2f"; // red
  const TEXT_COLOR = "#ffffff";

  const startBtn = document.getElementById("startBtn");
  const resetBtn = document.getElementById("resetBtn");
  const scoreValueEl = document.getElementById("scoreValue");
  let scaleX = 1;
  let scaleY = 1;
  let offsetX = 0;
  let offsetY = 0;

  function resizeCanvas() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    const MIN_CELL_SIZE = 15;
    const MAX_CELL_SIZE = 30;

    let scale = Math.max(
      MIN_CELL_SIZE,
      Math.min(MAX_CELL_SIZE, canvas.width / 80)
    );

    scaleX = scale;
    scaleY = scale;

    offsetX = canvas.width / 2;
    offsetY = canvas.height / 2;

    render();
  }

  // -----------------------------
  // Game State
  // -----------------------------
  let gameIntervalId = null;
  let serverTick = 0;
  let inputSeq = 0;
  let pendingInputs = [];

  const state = {
    snakes: {},
    isPaused: false,
    direction: { x: 1, y: 0 },
    nextDirection: { x: 1, y: 0 },
    apples: {},
    walls: [],
    growPending: 0,
    tickMs: 120,
    score: 0,
    socket: null,
    myPlayerId: null,
    myWorldX: 0,
    myWorldY: 0,
  };

  var myPlayerId;

  // -----------------------------
  // Utilities
  // -----------------------------
  function applyInput(snake, dir) {
    if ((!snake | !dir | snake.body, length == 0)) return;

    const head = snake.body[0];
    const newHead = { x: head.x + dir.x, y: head.y + dir.y };
    snake.body.shift(newHead);
    snake.body.pop();
  }

  // -----------------------------
  // Input handling
  // -----------------------------
  // Map of key -> direction vector
  const keyToDirection = {
    ArrowUp: { x: 0, y: -1 },
    ArrowDown: { x: 0, y: 1 },
    ArrowLeft: { x: -1, y: 0 },
    ArrowRight: { x: 1, y: 0 },
    w: { x: 0, y: -1 },
    s: { x: 0, y: 1 },
    a: { x: -1, y: 0 },
    d: { x: 1, y: 0 },
    W: { x: 0, y: -1 },
    S: { x: 0, y: 1 },
    A: { x: -1, y: 0 },
    D: { x: 1, y: 0 },
  };

  function handleKeydown(event) {
    if (event.code === "Space") {
      let data = {};
      if (!state.isPaused) {
        data = {
          type: "pause",
          roomId: gameId,
          playerId: myPlayerId,
        };
      } else {
        data = {
          type: "resume",
          roomId: gameId,
          playerId: myPlayerId,
        };
      }
      if (state.socket) state.socket.send(JSON.stringify(data));
      state.isPaused = !state.isPaused;
      return;
    }

    const newDir = keyToDirection[event.key];
    if (!newDir) return;

    inputSeq++;

    const mySnake = state.snakes[myPlayerId];

    if (mySnake) {
      applyInput(mySnake, newDir);
      render();
    }

    pendingInputs.push({ seq: inputSeq, dir: newDir });
    const data = {
      type: "move",
      room: gameId,
      playerId: myPlayerId,
      xOffset: keyToDirection[event.key].x,
      yOffset: keyToDirection[event.key].y,
      inputSeq: inputSeq,
    };
    if (state.socket != null) {
      state.socket.send(JSON.stringify(data));
      console.log(data);
      if (!newDir) return;
    }
  }

  // -----------------------------
  // Game logic
  // -----------------------------
  function resetGame() {
    state.snakes = {};
    state.direction = { x: 1, y: 0 };
    state.nextDirection = { x: 1, y: 0 };
    state.apple = { x: null, y: null };
    state.growPending = 0;
    state.score = 0;
    // Don't reset snakeColor - keep it from server
    updateScoreUI();
    clearBoard();
  }

  function startGame() {
    startBtn.style.display = "none";
    resetBtn.style.display = "none";
    document.addEventListener("keydown", handleKeydown);
    if (state.socket != null) {
      state.socket.send(JSON.stringify({ type: "start", room: gameId }));
    }
  }

  // -----------------------------
  // Rendering
  // -----------------------------
  function clearBoard() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
  }

  function drawCell(pos, color, isMySnake = false) {
    ctx.fillStyle = color;
    ctx.fillRect(pos.x, pos.y, scaleX, scaleY);

    if (isMySnake) {
      // Draw a black border around your snake
      ctx.lineWidth = 2;
      ctx.strokeStyle = "black";
      ctx.strokeRect(pos.x, pos.y, scaleX, scaleY);

      // Optional: add a shadow glow effect
      ctx.shadowColor = "rgba(0, 0, 0, 0.5)";
      ctx.shadowBlur = 4;
      ctx.shadowOffsetX = 0;
      ctx.shadowOffsetY = 0;
    } else {
      // Reset shadow for other snakes
      ctx.shadowColor = "transparent";
      ctx.shadowBlur = 0;
    }
  }

  function drawSnakes() {
    for (const [id, snake] of Object.entries(state.snakes)) {
      const color = snake.color || SNAKE_COLOR;
      for (const segment of snake.body) {
        const view = worldToView(segment);
        drawCell(view, color, snake.mySnake);
      }
    }
  }

  function drawApple() {
    for (const [id, apple] of Object.entries(state.apples)) {
      const view = worldToView(apple.pos);
      drawCell(view, apple.color);
    }
  }

  function drawWalls() {
    for (const [id, wall] of Object.entries(state.walls)) {
      const view = worldToView({ x: wall.x, y: wall.y });
      drawCell(view, "#ffffff");
    }
  }

  function render() {
    clearBoard();
    drawWalls();
    drawApple();
    drawSnakes();
  }

  function updateScoreUI() {
    if (scoreValueEl) scoreValueEl.textContent = String(state.score);
  }

  // -----------------------------
  // Button hooks
  // -----------------------------
  startBtn.addEventListener("click", () => {
    render();
    startGame();
  });

  resetBtn.addEventListener("click", () => {
    if (state.socket != null && state.socket.readyState === WebSocket.OPEN) {
      state.socket.send(JSON.stringify({ type: "reset", room: gameId }));
      resetGame();
    } else {
      console.warn("⚠️ WebSocket not open, retrying in 500ms...");
      setTimeout(() => {
        if (state.socket.readyState === WebSocket.OPEN) {
          state.socket.send(JSON.stringify({ type: "reset", room: gameId }));
        } else {
          console.error(
            "❌ Failed to send reset message — socket still closed"
          );
        }
      }, 500);
    }
  });

  document.getElementById("homeBtn").addEventListener("click", () => {
    window.location.href = "index.html";
  });

  // Initial render with start screen (not running)
  resetGame();
  clearBoard();
  render();
  window.addEventListener("resize", resizeCanvas);
  resizeCanvas(); // call once on startup

  const gameId = sessionStorage.getItem("gameId") || "1";
  if (!gameId.trim()) {
    alert("Please enter a game ID");
    return;
  }

  fetch(`https://snakeup.onrender.com/check-game/${gameId}`)
    .then((res) => {
      if (!res.ok) {
        window.location.href = "index.html";
        return Promise.reject("Game not valid");
      }
      
      const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
      const serverUrl = 'snakeup.onrender.com';

      const socket = new WebSocket(`${protocol}${serverUrl}/game/${gameId}`);
      state.socket = socket;
      setupSocket(socket);
    })
    .catch((err) => console.error(err));

  function worldToView(pos) {
    const dx = pos.x - state.myWorldX;
    const dy = pos.y - state.myWorldY;

    return {
      x: canvas.width / 2 + dx * scaleX,
      y: canvas.height / 2 + dy * scaleY,
    };
  }

  function setupSocket(socket) {
    socket.onopen = () => {
      console.log(`✅ Connected to WebSocket server at /game/${gameId}`);
      // You can send an initial message if needed
      socket.send(JSON.stringify({ type: "join", room: gameId }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);

      if (data.type == "player_init") {
        myPlayerId = data.playerId;

        console.log(
          `My id: ${myPlayerId}, Snake color: ${data.snakeColor}, Position: {${data.XPos}, ${data.YPos}}\n`
        );

        // Ensure there's an entry for my player in state.snakes (pre-lobby preview)
        // Use XPos/YPos from server if available (fallback to 0)
        const initX = typeof data.XPos === "number" ? data.XPos : 0;
        const initY = typeof data.YPos === "number" ? data.YPos : 0;

        state.myWorldX = data.XPos;
        state.myWorldY = data.YPos;
        state.snakes[myPlayerId] = {
          body: [{ x: initX, y: initY }],
          color: data.snakeColor || SNAKE_COLOR,
          mySnake: true,
        };

        // Render immediately so the player sees their snake in the lobby
        render();
      }

      if (data.type === "players_update") {
        const newSnakes = {};
        const newApples = {};
        const newWalls = {};

        console.log(data.walls);

        for (const p of data.players) {
          const body = p.body.map((seg) => ({
            x: Number(seg.x),
            y: Number(seg.y),
          }));
          const isMySnake = myPlayerId === p.playerId;

          if (isMySnake) {
            const mySnake = state.snakes[myPlayerId] || {};
            mySnake.body = body;
            state.myWorldX = mySnake.body[0].x;
            state.myWorldY = mySnake.body[0].y;
            mySnake.color = p.snakeColor;
            mySnake.mySnake = true;

            const lastProcessedInputSeq = p.lastProcessedInputSeq ?? 0;
            pendingInputs = pendingInputs.filter(
              (inp) => inp.seq > lastProcessedInputSeq
            );
            for (const inp of pendingInputs) {
              applyInput(mySnake, inp.dir);
            }

            newSnakes[myPlayerId] = mySnake;
          } else {
            newSnakes[p.playerId] = {
              body,
              color: p.snakeColor,
              mySnake: false,
            };
          }
        }
        for (const a of data.apples) {
          const pos = {
            x: Number(a.pos.X),
            y: Number(a.pos.Y),
          };

          newApples[a.appleId] = {
            id: a.appleId,
            color: a.appleColor,
            pos: pos,
          };
        }

        for (const w of data.walls) {
          const key = `${w.x},${w.y}`;
          newWalls[key] = { x: Number(w.x), y: Number(w.y) };
        }
        state.snakes = newSnakes;
        state.apples = newApples;
        state.walls = newWalls;
        render();
      }

      if (data.type == "player_died" && data.playerId == myPlayerId) {
        state.snakes = {};
        state.apple = {};
        clearBoard();
        // Grab elements
        const endScreen = document.getElementById("endScreen");
        const endMessage = document.getElementById("endMessage");

        if (endScreen && endMessage) {
          endScreen.style.display = "inline-block";
          endMessage.textContent = "You lose!";
        }
        render();
      }

      if (data.type === "remove_player") {
        for (const p of data.playerToRemove) {
          const id = p.playerId;

          // If it's me, then handle separately (optional)
          if (id === myPlayerId) {
            console.log("You have been removed from the game.");
            clearBoard();
            continue;
          }

          // Remove the player from the local snake state
          if (state.snakes[id]) {
            delete state.snakes[id];
          }
          console.log("Removed player:", id);
          render();
        }
      }

      if (data.type === "reset_game") {
        startBtn.style.display = "inline-block";
        resetBtn.style.display = "none";
        function clamp(v, min, max) {
          return Math.max(min, Math.min(max, v));
        }

        for (const p of data.players) {
          const x = clamp(p.x, 0, WORLD_COLS - 1);
          const y = clamp(p.y, 0, WORLD_ROWS - 1);
          const isMySnake = myPlayerId == p.playerId;
          state.snakes[p.playerId] = {
            body: [{ x, y }],
            color: p.snakeColor,
            mySnake: isMySnake,
          };
        }
        render();
      }

      if (data.type == "game_starting") {
        // set the snake states locally.
        serverTick = data.tickCount;
        if (serverTick === 0) {
          for (const p of data.players) {
            const x = p.x;
            const y = p.y;
            const isMySnake = myPlayerId == p.playerId;
            state.snakes[p.playerId] = {
              body: [{ x, y }],
              color: p.snakeColor,
              mySnake: isMySnake,
            };
          }
          console.log(
            "Starting position of snakes:",
            JSON.parse(JSON.stringify(state.snakes))
          );
          startBtn.style.display = "none";
          resetBtn.style.display = "none";
          document.addEventListener("keydown", handleKeydown);
          render();
        }
      }
      if (data.type == "game_over") {
        console.log("Ending game...");
        startBtn.style.display = "none";
        resetBtn.style.display = "inline-block";
        // Grab elements
        const endScreen = document.getElementById("endScreen");
        const endMessage = document.getElementById("endMessage");

        if (endScreen && endMessage) {
          endScreen.style.display = "inline-block";
          endMessage.textContent = "You win!";
        }
      }

      if (data.type == "update_score") {
        state.score = data.newScore;
        state.apple.x = data.apple.X;
        state.apple.y = data.apple.Y;
        if (data.playerId == myPlayerId) {
          updateScoreUI();
        }
        render();
      }

      if (data.type == "game_resetted") {
        state.score = 0;
        state.apple = { x: null, y: null };
        updateScoreUI();
        render();
      }
    };

    socket.onclose = () => {
      console.log("❌ Disconnected from WebSocket server");
    };

    socket.onerror = (error) => {
      console.error("⚠️ WebSocket error:", error);
    };
  }
})();
