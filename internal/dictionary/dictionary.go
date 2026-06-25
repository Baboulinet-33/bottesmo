package dictionary

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
)

var (
	wordsByLength = make(map[int][]string)
	fullWordSet   = make(map[string]bool)

	minLen, maxLen int
)

func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open dictionary %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := strings.TrimSpace(scanner.Text())
		w = strings.ToUpper(w)
		if w == "" {
			continue
		}
		l := len(w)
		wordsByLength[l] = append(wordsByLength[l], w)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dictionary: %w", err)
	}

	if len(wordsByLength) == 0 {
		return fmt.Errorf("dictionary is empty")
	}

	for l := range wordsByLength {
		sort.Strings(wordsByLength[l])
		if minLen == 0 || l < minLen {
			minLen = l
		}
		if l > maxLen {
			maxLen = l
		}
	}

	return nil
}

func LoadFull(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open full dictionary %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := strings.TrimSpace(scanner.Text())
		w = strings.ToUpper(w)
		if w == "" {
			continue
		}
		fullWordSet[w] = true
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading full dictionary: %w", err)
	}

	if len(fullWordSet) == 0 {
		return fmt.Errorf("full dictionary is empty")
	}

	return nil
}

func IsValid(word string) bool {
	return fullWordSet[strings.ToUpper(word)]
}

func WordsByLength(length int) []string {
	return wordsByLength[length]
}

func RandomWord(length int) (string, error) {
	words := wordsByLength[length]
	if len(words) == 0 {
		return "", fmt.Errorf("no words of length %d", length)
	}
	return words[rand.Intn(len(words))], nil
}

func DailyWord(length int, date string) (string, error) {
	words := wordsByLength[length]
	if len(words) == 0 {
		return "", fmt.Errorf("no words of length %d", length)
	}
	key := fmt.Sprintf("daily:%d:%s", length, date)
	h := sha256.Sum256([]byte(key))
	idx := int(h[0]) | (int(h[1]) << 8)
	return words[idx%len(words)], nil
}

func DailyLength(date string) int {
	h := sha256.Sum256([]byte("daily-length:" + date))
	sum := int(h[0])
	range_ := maxLen - minLen + 1
	return minLen + (sum % range_)
}

func Reset() {
	wordsByLength = make(map[int][]string)
	fullWordSet = make(map[string]bool)

	minLen = 0
	maxLen = 0
}

func MinLength() int {
	return minLen
}

func MaxLength() int {
	return maxLen
}
