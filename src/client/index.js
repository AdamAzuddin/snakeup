document.addEventListener('DOMContentLoaded', function() {
    const startBtn = document.getElementById('findBtn');

    startBtn.addEventListener('click', async function() {
        console.log("Start button clicked!");

        const gameId = document.getElementById('game-id').value.trim();
        if (!gameId) {
            alert("Please enter a game ID");
            return;
        }

        try {
            // 🔥 Make HTTP request to your backend join endpoint
            const response = await fetch(`http://localhost:42069/api/join/${encodeURIComponent(gameId)}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({}), // you can include player info here if needed
            });

            if (!response.ok) {
                const text = await response.text();
                console.error("❌ Failed to join game:", text);
                alert(`Failed to join game: ${text}`);
                return;
            }

            const result = await response.json();
            console.log("✅ Joined game successfully:", result);

            // 🔥 Optionally, you can store any player info the server returns
            if (result.playerId) {
                sessionStorage.setItem('playerId', result.playerId);
            }

            // Store gameId for the game page
            sessionStorage.setItem('gameId', gameId);

            // Navigate to game page
            window.location.href = "game.html";
        } catch (error) {
            console.error("⚠️ Error joining game:", error);
            document.getElementById('game-id').value = null
            alert(`Could not join the game with id ${gameId}.`);
        }
    });
});
