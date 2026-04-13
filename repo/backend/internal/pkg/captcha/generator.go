package captcha

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"encoding/base64"
	"math/big"
	mrand "math/rand"
	"sync"
	"time"
)

type Challenge struct {
	ID     string
	Answer string
	Image  string // base64 encoded PNG
}

type Store struct {
	mu         sync.RWMutex
	challenges map[string]*storedChallenge
}

type storedChallenge struct {
	answer    string
	expiresAt time.Time
}

func NewStore() *Store {
	s := &Store{
		challenges: make(map[string]*storedChallenge),
	}
	go s.cleanup()
	return s
}

func (s *Store) Generate() (*Challenge, error) {
	// Generate random arithmetic: a + b = ?
	a := randomInt(10, 99)
	b := randomInt(1, 20)
	answer := fmt.Sprintf("%d", a+b)
	text := fmt.Sprintf("%d + %d = ?", a, b)

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	img := renderCaptcha(text, 200, 60)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	s.mu.Lock()
	s.challenges[id] = &storedChallenge{
		answer:    answer,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	s.mu.Unlock()

	return &Challenge{
		ID:     id,
		Answer: answer,
		Image:  "data:image/png;base64," + b64,
	}, nil
}

func (s *Store) Verify(id, answer string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, ok := s.challenges[id]
	if !ok {
		return false
	}

	delete(s.challenges, id)

	if time.Now().After(ch.expiresAt) {
		return false
	}

	return ch.answer == answer
}

func (s *Store) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, ch := range s.challenges {
			if now.After(ch.expiresAt) {
				delete(s.challenges, id)
			}
		}
		s.mu.Unlock()
	}
}

func renderCaptcha(text string, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// White background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	// Add noise lines
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 5; i++ {
		x1, y1 := r.Intn(width), r.Intn(height)
		x2, y2 := r.Intn(width), r.Intn(height)
		drawLine(img, x1, y1, x2, y2, color.RGBA{R: uint8(r.Intn(200)), G: uint8(r.Intn(200)), B: uint8(r.Intn(200)), A: 255})
	}

	// Draw text characters as simple pixel blocks
	startX := 20
	for i, ch := range text {
		drawChar(img, startX+i*18, 15+r.Intn(10), ch, color.RGBA{R: uint8(r.Intn(100)), G: uint8(r.Intn(100)), B: uint8(r.Intn(100)), A: 255})
	}

	// Add noise dots
	for i := 0; i < 100; i++ {
		x, y := r.Intn(width), r.Intn(height)
		img.Set(x, y, color.RGBA{R: uint8(r.Intn(256)), G: uint8(r.Intn(256)), B: uint8(r.Intn(256)), A: 255})
	}

	return img
}

func drawChar(img *image.RGBA, x, y int, ch rune, clr color.Color) {
	// Simple 5x7 pixel font rendering for digits and basic chars
	patterns := map[rune][]string{
		'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		'2': {"01110", "10001", "00010", "00100", "01000", "10000", "11111"},
		'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		'6': {"01110", "10000", "11110", "10001", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'+': {"00000", "00100", "00100", "11111", "00100", "00100", "00000"},
		'=': {"00000", "00000", "11111", "00000", "11111", "00000", "00000"},
		'?': {"01110", "10001", "00001", "00110", "00100", "00000", "00100"},
		' ': {"00000", "00000", "00000", "00000", "00000", "00000", "00000"},
	}

	pattern, ok := patterns[ch]
	if !ok {
		return
	}

	scale := 3
	for row, line := range pattern {
		for col, bit := range line {
			if bit == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						px := x + col*scale + dx
						py := y + row*scale + dy
						if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
							img.Set(px, py, clr)
						}
					}
				}
			}
		}
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, clr color.Color) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, clr)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func randomInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
