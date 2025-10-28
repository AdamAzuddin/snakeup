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
  let serverTick = 0;
  let inputSeq = 0;
  let pendingInputs = [];

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

  function applyInput(snake, dir) {
    if ((!snake | !dir | snake.body, length == 0)) return;

    const head = snake.body[0];
    const newHead = { x: head.x + dir.x, y: head.y + dir.y };
    snake.body.shift(newHead);
    snake.body.pop();
  }

  function positionsEqual(a, b) {
    return a.x === b.x && a.y === b.y;
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

  function handleKeydown(event) {
    const newDir = keyToDirection[event.key];
    if (!newDir) return;

    inputSeq++;

    const mySnake = state.snakes[myPlayerId];

    if (mySnake) {
      applyInput(mySnake, newDir);
      console.log("Predicted move", inputSeq, newDir);
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

    socket.send(JSON.stringify(data));
    console.log(data);
    if (!newDir) return;
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
    console.log("start message sent");
  }

  // -----------------------------
  // Rendering
  // -----------------------------
  function clearBoard() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
  }

  function drawCell(pos, color, isMySnake) {
    const x = pos.x * GRID_SIZE;
    const y = pos.y * GRID_SIZE;

    // Draw the main snake body
    ctx.fillStyle = color;
    ctx.fillRect(x, y, GRID_SIZE, GRID_SIZE);

    if (isMySnake) {
      // Draw a black border around your snake
      ctx.lineWidth = 2;
      ctx.strokeStyle = "black";
      ctx.strokeRect(x, y, GRID_SIZE, GRID_SIZE);

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
          console.error(
            "❌ Failed to send reset message — socket still closed"
          );
        }
      }, 500);
    }
  });

  // Initial render with start screen (not running)
  resetGame();
  clearBoard();
  render();

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
      function clamp(v, min, max) {
        return Math.max(min, Math.min(max, v));
      }
      for (const p of data.players) {
        const x = clamp(p.x, 0, BOARD_COLS - 1);
        const y = clamp(p.y, 0, BOARD_ROWS - 1);
        const isMySnake = myPlayerId == p.playerId;

        if (isMySnake) {
          const mySnake = state.snakes[myPlayerId];
          mySnake.body = [{ x, y }];
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
          console.log("Reconciled with server, pending:", pendingInputs.length);
        } else {
          state.snakes[p.playerId] = {
            body: [{ x, y }],
            color: p.snakeColor,
            mySnake: false,
          };
        }
      }

      //console.log("🐍 Snakes state after update:", state.snakes);
      render();
    }

    if (data.type === "reset_game") {
      console.log("Resetting game...");
      startBtn.style.display = "inline-block";
      resetBtn.style.display = "none";
      function clamp(v, min, max) {
        return Math.max(min, Math.min(max, v));
      }

      for (const p of data.players) {
        const x = clamp(p.x, 0, BOARD_COLS - 1);
        const y = clamp(p.y, 0, BOARD_ROWS - 1);
        const isMySnake = myPlayerId == p.playerId;
        state.snakes[p.playerId] = {
          body: [{ x, y }],
          color: p.snakeColor,
          mySnake: isMySnake,
        };
      }

      console.log("🐍 Snakes state after update:", state.snakes);
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
          console.log(
            `[Client Tick ?] Player ${p.playerId} position: (${p.x}, ${p.y})`
          );
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
    }
  };

  socket.onclose = () => {
    console.log("❌ Disconnected from WebSocket server");
  };

  socket.onerror = (error) => {
    console.error("⚠️ WebSocket error:", error);
  };
})();
