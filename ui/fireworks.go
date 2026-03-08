package ui

import (
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fireworksDuration = 2 * time.Second
	fireworksFPS      = 30
	numParticles      = 60
	numRockets        = 5
)

var sparkChars = []string{"*", ".", "o", "+", "~", "·", "✦", "✧", "⊹", "⋆"}

var fireworkColors = []lipgloss.Color{
	lipgloss.Color("#FF6B6B"),
	lipgloss.Color("#4ECDC4"),
	lipgloss.Color("#FFE66D"),
	lipgloss.Color("#A8E6CF"),
	lipgloss.Color("#FF8B94"),
	lipgloss.Color("#DDA0DD"),
	lipgloss.Color("#98D8C8"),
	lipgloss.Color("#F7DC6F"),
}

type particle struct {
	x, y   float64
	vx, vy float64
	life   float64
	color  lipgloss.Color
	char   string
}

type fireworksModel struct {
	particles []particle
	width     int
	height    int
	startTime time.Time
	done      bool
}

type fireworksTickMsg time.Time
type fireworksDoneMsg struct{}

func newFireworks(width, height int) fireworksModel {
	m := fireworksModel{
		width:     width,
		height:    height,
		startTime: time.Now(),
	}
	m.spawnBurst(width, height)
	return m
}

func (m *fireworksModel) spawnBurst(width, height int) {
	for i := 0; i < numRockets; i++ {
		cx := float64(width/2) + (rand.Float64()-0.5)*float64(width/2)
		cy := float64(height/3) + (rand.Float64()-0.5)*float64(height/4)
		color := fireworkColors[rand.Intn(len(fireworkColors))]

		for j := 0; j < numParticles/numRockets; j++ {
			angle := rand.Float64() * 2 * math.Pi
			speed := rand.Float64()*3 + 1
			m.particles = append(m.particles, particle{
				x:     cx,
				y:     cy,
				vx:    math.Cos(angle) * speed,
				vy:    math.Sin(angle) * speed * 0.5, // squish vertically
				life:  0.5 + rand.Float64()*0.5,
				color: color,
				char:  sparkChars[rand.Intn(len(sparkChars))],
			})
		}
	}
}

func fireworksTickCmd() tea.Cmd {
	return tea.Tick(time.Second/fireworksFPS, func(t time.Time) tea.Msg {
		return fireworksTickMsg(t)
	})
}

func (m *fireworksModel) update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case fireworksTickMsg:
		if time.Since(m.startTime) > fireworksDuration {
			m.done = true
			return func() tea.Msg { return fireworksDoneMsg{} }
		}
		dt := 1.0 / float64(fireworksFPS)
		alive := make([]particle, 0, len(m.particles))
		for i := range m.particles {
			p := &m.particles[i]
			p.x += p.vx * dt * 8
			p.y += p.vy * dt * 4
			p.vy += dt * 2 // gravity
			p.life -= dt
			if p.life > 0 {
				alive = append(alive, *p)
			}
		}
		m.particles = alive

		// Spawn new bursts periodically
		elapsed := time.Since(m.startTime)
		if elapsed > 500*time.Millisecond && elapsed < fireworksDuration-500*time.Millisecond {
			if rand.Float64() < 0.15 {
				m.spawnBurst(m.width, m.height)
			}
		}

		return fireworksTickCmd()
	}
	return nil
}

func (m *fireworksModel) view(width, height int) string {
	if m.done || width == 0 || height == 0 {
		return ""
	}

	grid := make([][]rune, height)
	colors := make([][]lipgloss.Color, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		colors[i] = make([]lipgloss.Color, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	for _, p := range m.particles {
		x := int(p.x)
		y := int(p.y)
		if x >= 0 && x < width && y >= 0 && y < height {
			chars := []rune(p.char)
			if len(chars) > 0 {
				grid[y][x] = chars[0]
				colors[y][x] = p.color
			}
		}
	}

	// Render title in the center
	title := "✨ TUI Agent ✨"
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFE66D"))
	titleLen := len(title)
	titleY := height / 2
	titleX := (width - titleLen) / 2

	renderCell := func(y, x int) string {
		if grid[y][x] != ' ' {
			return lipgloss.NewStyle().Foreground(colors[y][x]).Render(string(grid[y][x]))
		}
		return " "
	}

	var result string
	for y := 0; y < height; y++ {
		line := ""
		if y == titleY {
			for x := 0; x < titleX; x++ {
				line += renderCell(y, x)
			}
			line += titleStyle.Render(title)
			for x := titleX + titleLen; x < width; x++ {
				line += renderCell(y, x)
			}
		} else {
			for x := 0; x < width; x++ {
				line += renderCell(y, x)
			}
		}
		result += line + "\n"
	}

	return result
}
