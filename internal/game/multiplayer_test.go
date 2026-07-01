package game

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tusmo/internal/dictionary"
)

func TestGenerateRoomCode(t *testing.T) {
	code := GenerateRoomCode()
	if len(code) != 6 {
		t.Errorf("expected code length 6, got %d", len(code))
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			if c < '2' || c > '9' {
				t.Errorf("unexpected character %c in room code", c)
			}
		}
	}
}

func TestNewMultiplayerRoom(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 5, "creator1", "Alice")
	if room.Code != "ABC123" {
		t.Errorf("expected Code ABC123, got %s", room.Code)
	}
	if room.Mode != "progressif" {
		t.Errorf("expected Mode progressif, got %s", room.Mode)
	}
	if room.WordCount != 5 {
		t.Errorf("expected WordCount 5, got %d", room.WordCount)
	}
	if room.CreatorID != "creator1" {
		t.Errorf("expected CreatorID creator1, got %s", room.CreatorID)
	}
	if room.State != "lobby" {
		t.Errorf("expected State lobby, got %s", room.State)
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}
	if room.MaxPlayers != 20 {
		t.Errorf("expected MaxPlayers 20, got %d", room.MaxPlayers)
	}
	creator := room.Players["creator1"]
	if creator == nil {
		t.Fatal("expected creator to be in Players")
	}
	if creator.Nickname != "Alice" {
		t.Errorf("expected nickname Alice, got %s", creator.Nickname)
	}
}

func TestAddPlayer(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 5, "creator1", "Alice")

	err := room.AddPlayer("player2", "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(room.Players))
	}
	if room.Players["player2"].Nickname != "Bob" {
		t.Errorf("expected nickname Bob, got %s", room.Players["player2"].Nickname)
	}

	err = room.AddPlayer("player2", "Bob")
	if err != nil {
		t.Fatalf("re-adding same player should not error: %v", err)
	}
}

func TestAddPlayerMaxPlayers(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 5, "creator1", "Alice")
	for i := 1; i < 20; i++ {
		room.AddPlayer("p"+string(rune('0'+i)), "Player")
	}
	err := room.AddPlayer("overflow", "Overflow")
	if err == nil {
		t.Error("expected error when room is full")
	}
}

func TestRemovePlayer(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 5, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")

	room.RemovePlayer("player2")
	if _, exists := room.Players["player2"]; exists {
		t.Error("expected player2 to be removed")
	}
	if len(room.Players) != 1 {
		t.Errorf("expected 1 player, got %d", len(room.Players))
	}

	room.RemovePlayer("nonexistent")
}

func TestAddPlayerMidGame_GeneratesWordSequence(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.StartGame()

	err := room.AddPlayer("player2", "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.WordSequence == nil {
		t.Fatal("expected room.WordSequence to be set")
	}
	if len(room.WordSequence) != 2 {
		t.Errorf("expected 2 words in WordSequence, got %d", len(room.WordSequence))
	}

	player := room.Players["player2"]
	if player == nil {
		t.Fatal("expected player2 to be in Players")
	}
	if len(player.WordGames) != 2 {
		t.Errorf("expected 2 games in WordGames, got %d", len(player.WordGames))
	}
	if player.CurrentWordIdx != 0 {
		t.Errorf("expected CurrentWordIdx 0, got %d", player.CurrentWordIdx)
	}
	if player.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}
}

func TestAddPlayerMidGame_MultipleWords(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 5, "creator1", "Alice")
	room.StartGame()

	err := room.AddPlayer("player2", "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(room.WordSequence) != 5 {
		t.Errorf("expected 5 words in room.WordSequence, got %d", len(room.WordSequence))
	}

	player := room.Players["player2"]
	if len(player.WordGames) != 5 {
		t.Errorf("expected 5 games, got %d", len(player.WordGames))
	}
	for i, g := range player.WordGames {
		if g == nil {
			t.Errorf("WordGames[%d] is nil", i)
		} else if g.Target != room.WordSequence[i] {
			t.Errorf("WordGames[%d].Target = %q, expected %q", i, g.Target, room.WordSequence[i])
		}
	}
}

func TestStartGame_NoPlayers_ReturnsError(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")
	room.RemovePlayer("creator1")

	err := room.StartGame()
	if err == nil {
		t.Error("expected error when starting game with no players")
	}
}

func TestStartGame(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")

	err := room.StartGame()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.State != "playing" {
		t.Errorf("expected State playing, got %s", room.State)
	}
	if room.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}
	if len(room.WordSequence) != 3 {
		t.Errorf("expected 3 words in room.WordSequence, got %d", len(room.WordSequence))
	}

	for id, player := range room.Players {
		if len(player.WordGames) != 3 {
			t.Errorf("player %s expected 3 games, got %d", id, len(player.WordGames))
		}
		for i, g := range player.WordGames {
			if g.Target != room.WordSequence[i] {
				t.Errorf("player %s WordGames[%d].Target = %q, expected %q", id, i, g.Target, room.WordSequence[i])
			}
		}
		if player.CurrentWordIdx != 0 {
			t.Errorf("player %s expected CurrentWordIdx 0, got %d", id, player.CurrentWordIdx)
		}
		if player.StartTime.IsZero() {
			t.Errorf("player %s expected StartTime to be set", id)
		}
	}
}

func TestStartGameAlreadyStarted(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")
	room.StartGame()
	err := room.StartGame()
	if err == nil {
		t.Error("expected error when starting already started game")
	}
}

func TestRestartGame(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")

	if err := room.StartGame(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := room.RestartGame(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.State != "lobby" {
		t.Errorf("expected State lobby, got %s", room.State)
	}
	if room.WordSequence != nil {
		t.Errorf("expected WordSequence nil, got %v", room.WordSequence)
	}
	if !room.StartTime.IsZero() {
		t.Error("expected StartTime to be zero")
	}

	for id, player := range room.Players {
		if player.WordGames != nil {
			t.Errorf("player %s expected WordGames nil, got %v", id, player.WordGames)
		}
		if player.CurrentWordIdx != 0 {
			t.Errorf("player %s expected CurrentWordIdx 0, got %d", id, player.CurrentWordIdx)
		}
		if !player.StartTime.IsZero() {
			t.Errorf("player %s expected StartTime zero", id)
		}
		if !player.CompletedTime.IsZero() {
			t.Errorf("player %s expected CompletedTime zero", id)
		}
		if player.Failed {
			t.Errorf("player %s expected Failed false", id)
		}
		if player.Finished {
			t.Errorf("player %s expected Finished false", id)
		}
	}

	if room.CreatorID != "creator1" {
		t.Errorf("expected CreatorID preserved, got %s", room.CreatorID)
	}
	if len(room.Players) != 2 {
		t.Errorf("expected 2 players preserved, got %d", len(room.Players))
	}
}

func TestRestartGame_NotPlaying(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")

	err := room.RestartGame()
	if err == nil {
		t.Error("expected error when restarting a game in lobby state")
	}
}

func TestRestartGame_MultiplePlayers(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 3, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")
	room.AddPlayer("player3", "Charlie")

	if err := room.StartGame(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	room.Players["creator1"].CurrentWordIdx = 2
	room.Players["creator1"].Finished = true
	room.Players["creator1"].Failed = false
	room.Players["creator1"].CompletedTime = time.Now()

	room.Players["player2"].CurrentWordIdx = 1
	room.Players["player2"].Failed = true
	room.Players["player2"].Finished = true
	room.Players["player2"].CompletedTime = time.Now()

	if err := room.RestartGame(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for id, player := range room.Players {
		if player.WordGames != nil {
			t.Errorf("player %s expected WordGames nil", id)
		}
		if player.CurrentWordIdx != 0 {
			t.Errorf("player %s expected CurrentWordIdx 0, got %d", id, player.CurrentWordIdx)
		}
		if player.Failed {
			t.Errorf("player %s expected Failed false", id)
		}
		if player.Finished {
			t.Errorf("player %s expected Finished false", id)
		}
		if !player.StartTime.IsZero() {
			t.Errorf("player %s expected StartTime zero", id)
		}
		if !player.CompletedTime.IsZero() {
			t.Errorf("player %s expected CompletedTime zero", id)
		}
	}

	if len(room.Players) != 3 {
		t.Errorf("expected 3 players preserved, got %d", len(room.Players))
	}
}

func TestProcessGuessCorrect(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.StartGame()

	player := room.Players["creator1"]
	target := player.WordGames[0].Target

	result, err := room.ProcessGuess("creator1", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.WordFinished {
		t.Error("expected WordFinished")
	}
	if !result.WordWon {
		t.Error("expected WordWon")
	}
	if result.PlayerFinished {
		t.Error("expected PlayerFinished false (has more words)")
	}
	if result.PlayerFailed {
		t.Error("expected PlayerFailed false")
	}
	if player.CurrentWordIdx != 1 {
		t.Errorf("expected CurrentWordIdx 1, got %d", player.CurrentWordIdx)
	}
}

func TestProcessGuessLastWord(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 1, "creator1", "Alice")
	room.StartGame()

	player := room.Players["creator1"]
	target := player.WordGames[0].Target

	result, err := room.ProcessGuess("creator1", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.PlayerFinished {
		t.Error("expected PlayerFinished")
	}
	if result.PlayerFailed {
		t.Error("expected PlayerFailed false")
	}
	if !player.Finished {
		t.Error("expected player Finished")
	}
	if player.Failed {
		t.Error("expected player Failed false")
	}
	if player.CompletedTime.IsZero() {
		t.Error("expected CompletedTime to be set")
	}
}

func TestProcessGuessFail(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.StartGame()

	player := room.Players["creator1"]
	target := player.WordGames[0].Target
	firstLetter := string(target[0])
	wrongGuesses := []string{
		firstLetter + "BBBBB",
		firstLetter + "CCCCC",
		firstLetter + "DDDDD",
		firstLetter + "EEEEE",
		firstLetter + "FFFFF",
		firstLetter + "GGGGG",
	}

	for i, guess := range wrongGuesses {
		result, err := room.ProcessGuess("creator1", guess)
		if err == nil {
			if i < 5 {
				if result.WordFinished {
					t.Fatalf("guess %d: expected WordFinished false", i)
				}
			} else {
				if !result.PlayerFinished {
					t.Fatalf("guess %d: expected PlayerFinished", i)
				}
				if !result.PlayerFailed {
					t.Error("expected PlayerFailed")
				}
			}
		}
	}

	if !player.Finished {
		t.Error("expected player Finished after 6 failed attempts")
	}
	if !player.Failed {
		t.Error("expected player Failed")
	}
}

func TestProcessGuessAfterFinish(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 1, "creator1", "Alice")
	room.StartGame()

	player := room.Players["creator1"]
	player.Finished = true

	_, err := room.ProcessGuess("creator1", "TEST")
	if err == nil {
		t.Error("expected error when player already finished")
	}
}

func TestGetRankings(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")
	room.AddPlayer("player3", "Charlie")
	room.StartGame()

	rankings := room.GetRankings()
	if len(rankings) != 3 {
		t.Errorf("expected 3 rankings, got %d", len(rankings))
	}

	for _, r := range rankings {
		if r.Finished {
			t.Error("expected no one finished yet")
		}
	}

	player1 := room.Players["creator1"]
	for i := 0; i < room.WordCount; i++ {
		target := player1.WordGames[i].Target
		_, err := room.ProcessGuess("creator1", target)
		if err != nil {
			t.Fatalf("unexpected error guessing word %d: %v", i, err)
		}
	}

	rankings = room.GetRankings()
	if !rankings[0].Finished {
		t.Error("expected first ranking to be finished")
	}

	aliceRanking := rankings[0]
	if len(aliceRanking.WordResults) != room.WordCount {
		t.Errorf("expected %d WordResults, got %d", room.WordCount, len(aliceRanking.WordResults))
	}
	for i, wr := range aliceRanking.WordResults {
		expected := player1.WordGames[i]
		if wr.Target != expected.Target {
			t.Errorf("word %d: expected target %q, got %q", i, expected.Target, wr.Target)
		}
		if len(wr.Attempts) != len(expected.Attempts) {
			t.Errorf("word %d: expected %d attempts, got %d", i, len(expected.Attempts), len(wr.Attempts))
		}
		if len(wr.Results) != len(expected.Attempts) {
			t.Errorf("word %d: expected %d result rows, got %d", i, len(expected.Attempts), len(wr.Results))
		}
	}
}

func TestIsRoomFinished(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.AddPlayer("player2", "Bob")
	room.StartGame()

	if room.IsRoomFinished() {
		t.Error("expected room not finished")
	}

	room.Players["creator1"].Finished = true
	room.Players["player2"].Finished = true

	if !room.IsRoomFinished() {
		t.Error("expected room finished")
	}
}

func TestComputeAttemptResults(t *testing.T) {
	target := "ABRITE"
	attempts := []string{"ABIDES"}
	results, err := ComputeAttemptResults(target, attempts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0]) != 6 {
		t.Fatalf("expected 6 letter results, got %d", len(results[0]))
	}
	if results[0][0].Status != StatusCorrect {
		t.Errorf("expected first letter correct")
	}
	if results[0][1].Status != StatusCorrect {
		t.Errorf("expected second letter correct")
	}
}

func TestMidGameJoin_StateRestore(t *testing.T) {
	room := NewMultiplayerRoom("ABC123", "progressif", 2, "creator1", "Alice")
	room.StartGame()

	player := room.Players["creator1"]
	target0 := player.WordGames[0].Target
	room.ProcessGuess("creator1", target0)

	if player.CurrentWordIdx != 1 {
		t.Errorf("expected CurrentWordIdx 1, got %d", player.CurrentWordIdx)
	}
	if len(room.WordSequence) != 2 {
		t.Errorf("expected 2 words, got %d", len(room.WordSequence))
	}
	if !player.WordGames[0].GameOver {
		t.Errorf("expected first word game to be over")
	}
}

func TestGenerateWordSequence_Progressive(t *testing.T) {
	dictPath := setupTestDict(t)
	dictionary.Reset()
	if err := dictionary.Load(dictPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	seq := GenerateWordSequence("progressif", 7)
	if len(seq) != 7 {
		t.Fatalf("expected 7 words, got %d", len(seq))
	}
	expectedLens := []int{6, 7, 8, 9, 10, 6, 7}
	for i, word := range seq {
		if len(word) != expectedLens[i] {
			t.Errorf("word %d: expected length %d, got %d (%q)", i, expectedLens[i], len(word), word)
		}
	}
}

func TestGenerateWordSequence_Aleatoire(t *testing.T) {
	dictPath := setupTestDict(t)
	dictionary.Reset()
	if err := dictionary.Load(dictPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	seq := GenerateWordSequence("aleatoire", 5)
	if len(seq) != 5 {
		t.Fatalf("expected 5 words, got %d", len(seq))
	}
	for i, word := range seq {
		if len(word) < 6 || len(word) > 10 {
			t.Errorf("word %d: expected length between 6 and 10, got %d (%q)", i, len(word), word)
		}
	}
}

func TestGenerateWordSequence_Count1(t *testing.T) {
	dictPath := setupTestDict(t)
	dictionary.Reset()
	if err := dictionary.Load(dictPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	seq := GenerateWordSequence("progressif", 1)
	if len(seq) != 1 {
		t.Fatalf("expected 1 word, got %d", len(seq))
	}
	if len(seq[0]) != 6 {
		t.Errorf("expected length 6 for first word, got %d", len(seq[0]))
	}
}

func TestGenerateWordSequence_Count10(t *testing.T) {
	dictPath := setupTestDict(t)
	dictionary.Reset()
	if err := dictionary.Load(dictPath); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	seq := GenerateWordSequence("progressif", 10)
	if len(seq) != 10 {
		t.Fatalf("expected 10 words, got %d", len(seq))
	}
	expectedLens := []int{6, 7, 8, 9, 10, 6, 7, 8, 9, 10}
	for i, word := range seq {
		if len(word) != expectedLens[i] {
			t.Errorf("word %d: expected length %d, got %d (%q)", i, expectedLens[i], len(word), word)
		}
	}
}

func setupTestDict(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "words.txt")
	content := "ABRITE\nACCORD\nACTION\nABANDON\nABATTRE\nABDOMEN\nABEILLE\nABONNER\nABORDEE\n"
	content += "ABRITER\nABROGER\nABRUPTS\nABSCONS\nABSENCE\nABSOLUE\nABSOUDRE\nABSURDE\nABUSIVE\n"
	content += "BANALITE\nBANANE\nBANANIER\nBANCAL\nBANDEAU\nBANLIEUE\n"
	content += "BANNIERE\nBANQUET\nBANQUETTE\nBANQUIER\nBAPTEME\nBAPTISER\n"
	content += "BARBANT\nBARBARE\nBARBE\nBARBEAU\nBARBIER\nBARBOTER\n"
	content += "BATAILLE\nBATAILLON\nBATEAU\nBATIMENT\nBATISSE\nBATON\n"
	content += "BATTANT\nBATTERIE\nBATTEUR\nBATTRE\nBAVARD\nBAVARDAGE\n"
	content += "BAVARDER\nBAVURE\nBEAUCOUP\nBEAUTE\nBEBE\nBECANE\n"
	content += "BECASSE\nBECHE\nBECHER\nBECOTER\nBEDON\nBEGAIE\n"
	content += "BEGONIA\nBEGUIN\nBEIGE\nBEIGNET\nBELETTE\nBELGE\n"
	content += "BELIER\nBELLE\nBENEFICE\nBENIR\nBENIT\nBENOIT\n"
	content += "BERCAIL\nBERCEAU\nBERCER\nBERET\nBERGE\nBERGER\n"
	content += "BERLINE\nBERLUE\nBERNARD\nBERNE\nBERNER\nBESACE\n"
	content += "BESOGNE\nBESOIN\nBETAIL\nBETE\nBETISE\nBETON\n"
	content += "BEURRE\nBEURRER\nBEURRIER\nBEVUE\nBIAIS\nBIBELOT\n"
	content += "BIBERON\nBIBLE\nBICHE\nBICHON\nBICYCLE\nBICYCLETTE\n"
	content += "BIDET\nBIDON\nBIEN\nBIENFAIT\nBIERE\nBIGAME\n"
	content += "BIGOT\nBIJOU\nBIJOUTIER\nBILAN\nBILINGUE\nBILLE\n"
	content += "BILLET\nBILLION\nBILLON\nBINAIRE\nBINETTE\nBINOCLE\n"
	content += "BINOME\nBIOGRAPHIE\nBIOLOGIE\nBIOLOGIQUE\nBIOMASSE\nBIOPHYSIQUE\n"
	content += "BISCUIT\nBISCUITERIE\nBISE\nBISEAU\nBISON\nBISQUE\n"
	content += "BISTOURI\nBISTRO\nBISTROT\nBITUME\nBITUMEUX\nBIVOUAC\n"
	content += "BIZARRE\nBIZARRERIE\nBIZUT\nBLAGUE\nBLAGUER\nBLAGUEUR\n"
	content += "BLAIREAU\nBLAMER\nBLANC\nBLANCHEUR\nBLANCHIR\nBLANCHISSEUR\n"
	content += "BLASE\nBLASER\nBLASPHME\nBLEME\nBLEMIR\nBLESSE\n"
	content += "BLESSER\nBLESSURE\nBLETTE\nBLEU\nBLEUET\nBLINDAGE\n"
	content += "BLINDER\nBLOC\nBLOCAGE\nBLOND\nBLONDEUR\nBLOQUER\n"
	content += "BLOTTIR\nBLOUSE\nBLUES\nBLUETTE\nBLUFF\nBLUFFER\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
