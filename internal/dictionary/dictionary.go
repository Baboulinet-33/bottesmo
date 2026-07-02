package dictionary

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	wordsByLength = make(map[int][]string)
	fullWordSet   = make(map[string]bool)

	minLen, maxLen int
)

type LoadResult struct {
	Version   string
	WordCount int
	SHA256    [32]byte
}

func parseVersionLine(line string) (string, bool) {
	if strings.HasPrefix(line, "DICT_VERSION=") {
		return strings.TrimPrefix(line, "DICT_VERSION="), true
	}
	return "", false
}

func processWord(w string) string {
	w = strings.TrimSpace(w)
	return strings.ToUpper(w)
}

func forEachWord(data []byte, fn func(string)) LoadResult {
	sha := sha256.Sum256(data)
	lines := strings.Split(string(data), "\n")

	result := LoadResult{SHA256: sha}

	startIdx := 0
	version, isVersion := parseVersionLine(strings.TrimSpace(lines[0]))
	if isVersion {
		result.Version = version
		startIdx = 1
	}

	for i := startIdx; i < len(lines); i++ {
		w := processWord(lines[i])
		if w == "" {
			continue
		}
		fn(w)
		result.WordCount++
	}

	return result
}

func LoadFromBytes(data []byte) (LoadResult, error) {
	result := forEachWord(data, func(w string) {
		l := len(w)
		wordsByLength[l] = append(wordsByLength[l], w)
	})

	if result.WordCount == 0 {
		return result, fmt.Errorf("dictionary is empty")
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

	return result, nil
}

func LoadFullFromBytes(data []byte) (LoadResult, error) {
	result := forEachWord(data, func(w string) {
		fullWordSet[w] = true
	})

	if result.WordCount == 0 {
		return result, fmt.Errorf("full dictionary is empty")
	}

	return result, nil
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func resolveSource(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := httpClient.Get(source)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch dictionary %s: %w", source, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d fetching dictionary %s", resp.StatusCode, source)
		}

		limited := io.LimitReader(resp.Body, 100*1024*1024)
		data, err := io.ReadAll(limited)
		if err != nil {
			return nil, fmt.Errorf("cannot read dictionary response %s: %w", source, err)
		}
		return data, nil
	}
	clean := filepath.Clean(source)
	if strings.Contains(clean, "..") {
		return nil, fmt.Errorf("invalid dictionary source: %s", source)
	}
	return os.ReadFile(clean)
}

func LoadFromSource(source string) (LoadResult, error) {
	data, err := resolveSource(source)
	if err != nil {
		return LoadResult{}, fmt.Errorf("cannot load dictionary %s: %w", source, err)
	}

	result, err := LoadFromBytes(data)
	if err != nil {
		return result, err
	}

	log.Printf("Dictionary loaded: version=%s words=%d sha256=%x", result.Version, result.WordCount, result.SHA256)
	return result, nil
}

func LoadFullFromSource(source string) (LoadResult, error) {
	data, err := resolveSource(source)
	if err != nil {
		return LoadResult{}, fmt.Errorf("cannot load full dictionary %s: %w", source, err)
	}

	result, err := LoadFullFromBytes(data)
	if err != nil {
		return result, err
	}

	log.Printf("Full dictionary loaded: version=%s words=%d sha256=%x", result.Version, result.WordCount, result.SHA256)
	return result, nil
}

func Load(path string) error {
	_, err := LoadFromSource(path)
	return err
}

func LoadFull(path string) error {
	_, err := LoadFullFromSource(path)
	return err
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
