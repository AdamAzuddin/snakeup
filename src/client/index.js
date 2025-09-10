document.addEventListener('DOMContentLoaded', function() {
    const startBtn = document.getElementById('findBtn');
    
    startBtn.addEventListener('click', function() {
        console.log("Start button clicked!");
        
        const gameId = document.getElementById('game-id').value;
        if (!gameId.trim()) {
            alert("Please enter a game ID");
            return;
        }
        
        // Store gameId for the game page
        sessionStorage.setItem('gameId', gameId);
        
        // Navigate to game page
        window.location.href = "game.html";
    });
});