// SnakeUp - Simple Snake game using HTML Canvas
// The game is modular and easy to extend.

(() => {
  // -----------------------------
  // Constants and configuration
  // -----------------------------
  const canvas = document.getElementById("gameCanvas");
  const ctx = canvas.getContext("2d");

  const GRID_SIZE = 20; // width/height of one grid cell in pixels
  const BOARD_COLS = Math.floor(canvas.width / GRID_SIZE);
  const BOARD_ROWS = Math.floor(canvas.height / GRID_SIZE);

  const SNAKE_COLOR = "#1976d2"; // blue (fallback)
  const APPLE_COLOR = "#d32f2f"; // red
  const TEXT_COLOR = "#ffffff";

  const startBtn = document.getElementById("startBtn");
  const resetBtn = document.getElementById("resetBtn");
  const scoreValueEl = document.getElementById("scoreValue");

  // -----------------------------
  // Game State
  // -----------------------------
  let gameIntervalId = null;

  const state = {
    snakes: {},
    direction: { x: 1, y: 0 },
    nextDirection: { x: 1, y: 0 },
    apple: spawnApple([{ x: 5, y: 5 }]), // pass initial snake
    growPending: 0,
    tickMs: 120,
    score: 0,
    socket: null,
    myPlayerId: null,
  };

  var myPlayerId;

  // -----------------------------
  // Utilities
  // -----------------------------
  function randomInt(minInclusive, maxInclusive) {
    return (
      Math.floor(Math.random() * (maxInclusive - minInclusive + 1)) +
      minInclusive
    );
  }

  function positionsEqual(a, b) {
    return a.x === b.x && a.y === b.y;
  }

  function isInsideBoard(pos) {
    return pos.x >= 0 && pos.x < BOARD_COLS && pos.y >= 0 && pos.y < BOARD_ROWS;
  }

  function spawnApple(snake) {
    let pos;
    do {
      pos = {
        x: randomInt(0, BOARD_COLS - 1),
        y: randomInt(0, BOARD_ROWS - 1),
      };
    } while (snake.some((seg) => positionsEqual(seg, pos)));
    return pos;
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

  function areOpposite(d1, d2) {
    return d1.x === -d2.x && d1.y === -d2.y;
  }

  function handleKeydown(event) {
    const newDir = keyToDirection[event.key];
    if (!newDir) return;
    // Prevent reversing direction in the same tick
    if (!areOpposite(newDir, state.direction)) {
      state.nextDirection = newDir;
    }
  }

  // -----------------------------
  // Game logic
  // -----------------------------
  function resetGame() {
    state.snakes = {};
    state.direction = { x: 1, y: 0 };
    state.nextDirection = { x: 1, y: 0 };
    state.apple = { x: 0, y: 0 };
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
    socket.send(JSON.stringify({ type: "start", room: gameId }));
    console.log("start message sent")
  }

  function startGameForEveryone() {
    startBtn.style.display = "none";
    resetBtn.style.display = "none";
    document.addEventListener("keydown", handleKeydown);
  }

  // -----------------------------
  // Rendering
  // -----------------------------
  function clearBoard() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
  }

  function drawCell(pos, color) {
    ctx.fillStyle = color;
    ctx.fillRect(pos.x * GRID_SIZE, pos.y * GRID_SIZE, GRID_SIZE, GRID_SIZE);
  }

  function drawSnakes() {
    for (const [id, snake] of Object.entries(state.snakes)) {
      const color = snake.color || SNAKE_COLOR;
      for (const segment of snake.body) {
        drawCell(segment, color);
      }
    }
  }

  function drawApple() {
    drawCell(state.apple, APPLE_COLOR);
  }

  function drawLoseText() {
    ctx.fillStyle = TEXT_COLOR;
    ctx.font =
      "bold 32px system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("You lose!", canvas.width / 2, canvas.height / 2);
  }

  function render() {
    clearBoard();
    //drawApple();
    drawSnakes();
  }

  function updateScoreUI() {
    if (scoreValueEl) scoreValueEl.textContent = String(state.score);
  }

  // -----------------------------
  // Button hooks
  // -----------------------------
  startBtn.addEventListener("click", () => {
    console.log("Start button clicked");
    render();
    startGame();
  });

  resetBtn.addEventListener("click", () => {
  console.log("Reset button clicked");

  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "reset", room: gameId }));
    console.log("✅ Reset message sent");
  } else {
    console.warn("⚠️ WebSocket not open, retrying in 500ms...");
    setTimeout(() => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "reset", room: gameId }));
        console.log("✅ Reset message sent after retry");
      } else {
        console.error("❌ Failed to send reset message — socket still closed");
      }
    }, 500);
  }
});


  // Initial render with start screen (not running)
  resetGame();
  clearBoard();

  // connect with websocket
  const gameId = sessionStorage.getItem("gameId") || "1";
  if (!gameId.trim()) {
    alert("Please enter a game ID");
    return;
  }

  // start a websocket connection
  const socket = new WebSocket(`ws://localhost:42069/game/${gameId}`);

  socket.onopen = () => {
    console.log(`✅ Connected to WebSocket server at /game/${gameId}`);
    // You can send an initial message if needed
    socket.send(JSON.stringify({ type: "join", room: gameId }));
  };

  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    //console.log("📩 Message from server:", data);

    if (data.type == "player_init") {
      myPlayerId = data.playerId;
      console.log(
        `My id: ${myPlayerId}, Snake color: ${data.snakeColor}, Position: {${data.XPos}, ${data.YPos}}\n`
      );
    }
    if (data.type === "players_update") {
      function clamp(v, min, max) {
        return Math.max(min, Math.min(max, v));
      }

      for (const p of data.players) {
        const x = clamp(p.x, 0, BOARD_COLS - 1);
        const y = clamp(p.y, 0, BOARD_ROWS - 1);
        state.snakes[p.playerId] = {
          body: [{ x, y }],
          color: p.snakeColor,
        };
      }

      console.log("🐍 Snakes state after update:", state.snakes);
      render();
    }

    if (data.type === "reset_game") {
      console.log("Resetting game...");
      startBtn.style.display = "inline-block";
      resetBtn.style.display = "none";
      function clamp(v, min, max) {
        return Math.max(min, Math.min(max, v))    ;
      }

      for (const p of data.players) {
        const x = clamp(p.x, 0, BOARD_COLS - 1);
        const y = clamp(p.y, 0, BOARD_ROWS - 1);
        state.snakes[p.playerId] = {
          body: [{ x, y }],
          color: p.snakeColor,
        };
      }

      console.log("🐍 Snakes state after update:", state.snakes);
      render();
    }

    if (data.type == "game_starting") {
      startGameForEveryone();
    }
    if (data.type == "game_over") {
      console.log("Ending game...");
      startBtn.style.display = "none";
      resetBtn.style.display = "inline-block";
    }
  };

  socket.onclose = () => {
    console.log("❌ Disconnected from WebSocket server");
  };

  socket.onerror = (error) => {
    console.error("⚠️ WebSocket error:", error);
  };

  
})();
