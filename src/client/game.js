// SnakeUp - Simple Snake game using HTML Canvas
// The game is modular and easy to extend.

(() => {
  // -----------------------------
  // Constants and configuration
  // -----------------------------
  const canvas = document.getElementById("gameCanvas");
  const ctx = canvas.getContext("2d");

  const WORLD_COLS = 178 / 2;
  const WORLD_ROWS = 100 / 2;

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
    console.log("Canvas resized");

    // Use a single scale factor to keep cells square
    const scale = Math.min(
      canvas.width / WORLD_COLS,
      canvas.height / WORLD_ROWS
    );

    // Save scale for rendering
    scaleX = scale;
    scaleY = scale;

    // Center the game board
    offsetX = (canvas.width - WORLD_COLS * scale) / 2;
    offsetY = (canvas.height - WORLD_ROWS * scale) / 2;

    clearBoard();
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
    direction: { x: 1, y: 0 },
    nextDirection: { x: 1, y: 0 },
    apple: { x: null, y: null },
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
    const x = offsetX + pos.x * scaleX;
    const y = offsetY + pos.y * scaleY;
    // Draw the main snake bodymy
    ctx.fillStyle = color;
    ctx.fillRect(x, y, scaleX, scaleY);

    if (isMySnake) {
      // Draw a black border around your snake
      ctx.lineWidth = 2;
      ctx.strokeStyle = "black";
      ctx.strokeRect(x, y, scaleX, scaleY);

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
        drawCell(segment, color, snake.mySnake);
      }
    }
  }

  function drawApple() {
    if (state.apple.x != null && state.apple.y != null) {
      drawCell(state.apple, APPLE_COLOR);
    }
  }

  function render() {
    clearBoard();
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

  fetch(`http://localhost:42069/check-game/${gameId}`)
    .then((res) => {
      if (!res.ok) {
        window.location.href = "index.html";
        return Promise.reject("Game not valid");
      }
      const socket = new WebSocket(`ws://localhost:42069/game/${gameId}`);
      state.socket = socket;
      setupSocket(socket);
    })
    .catch((err) => console.error(err));

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

        state.snakes[myPlayerId] = {
          body: [{ x: initX, y: initY }],
          color: data.snakeColor || SNAKE_COLOR,
          mySnake: true,
        };

        // Render immediately so the player sees their snake in the lobby
        render();
      }

      if (data.type === "players_update") {
        for (const p of data.players) {
          const body = p.body.map((seg) => ({
            x: Number(seg.x),
            y: Number(seg.y),
          }));
          const isMySnake = myPlayerId == p.playerId;

          if (isMySnake) {
            const mySnake = state.snakes[myPlayerId];
            mySnake.body = body;
            mySnake.color = p.snakeColor;
            mySnake.mySnake = true;

            const lastProcessedInputSeq = p.lastProcessedInputSeq ?? 0;
            pendingInputs = pendingInputs.filter(
              (inp) => inp.seq > lastProcessedInputSeq
            );

            for (const inp of pendingInputs) {
              applyInput(mySnake, inp.dir);
            }
            state.snakes[myPlayerId] = mySnake;
          } else {
            state.snakes[p.playerId] = {
              body: body,
              color: p.snakeColor,
              mySnake: false,
            };
          }
        }

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
        console.log(data.apple);
        state.apple.x = data.apple.X;
        state.apple.y = data.apple.Y;
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
