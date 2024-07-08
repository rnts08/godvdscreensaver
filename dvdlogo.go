package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

//go:embed dvd-logo.png
var logoImageData []byte // Embedded the logo image

const (
	screenWidth       = 800
	screenHeight      = 600
	logoWidth         = 120
	logoStartVelocity = 2
	logoMaxVelocity   = 3
	cornerTolerance   = 5
	nudgeAmount       = 0.5
)

type Game struct {
	logoX      float64
	logoY      float64
	velocityX  float64
	velocityY  float64
	cornerHits int
	startTime  time.Time
	logoImage  *ebiten.Image
	logoHeight float64
	hitCorner  bool
	paused     bool
	terminated bool
	keyState   map[ebiten.Key]bool
}

func (g *Game) Update() error {
	// Handle key press events
	g.handleKeyPresses()

	if g.paused {
		if g.terminated {
			return ebiten.Termination
		}
		return nil
	}

	g.logoX += g.velocityX
	g.logoY += g.velocityY
	g.hitCorner = false

	// Check for collision with window borders
	if g.logoX < 0 {
		g.logoX = 0
		g.velocityX = -g.velocityX
	}
	if g.logoX+logoWidth > screenWidth {
		g.logoX = screenWidth - logoWidth
		g.velocityX = -g.velocityX
	}
	if g.logoY < 0 {
		g.logoY = 0
		g.velocityY = -g.velocityY
	}
	if g.logoY+g.logoHeight > screenHeight {
		g.logoY = screenHeight - g.logoHeight
		g.velocityY = -g.velocityY
	}

	// Check if the logo touches the corner
	if g.logoX < cornerTolerance || g.logoX > screenWidth-logoWidth-cornerTolerance {
		if g.logoY < cornerTolerance || g.logoY > screenHeight-g.logoHeight-cornerTolerance {
			g.cornerHits++
			g.hitCorner = true
		}
	}

	// Adjust velocity based on mouse input
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		dx := float64(x) - (g.logoX + logoWidth/2)
		dy := float64(y) - (g.logoY + g.logoHeight/2)
		g.velocityX += dx * nudgeAmount / 1000
		g.velocityY += dy * nudgeAmount / 1000

		// Clamp velocity to logoMaxVelocity
		if math.Abs(g.velocityX) > logoMaxVelocity {
			g.velocityX = math.Copysign(logoMaxVelocity, g.velocityX)
		}
		if math.Abs(g.velocityY) > logoMaxVelocity {
			g.velocityY = math.Copysign(logoMaxVelocity, g.velocityY)
		}
	}

	return nil
}

func (g *Game) handleKeyPresses() {
	// Check for escape key press to toggle pause state
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		if !g.keyState[ebiten.KeyEscape] {
			g.paused = !g.paused
		}
		g.keyState[ebiten.KeyEscape] = true
	} else {
		g.keyState[ebiten.KeyEscape] = false
	}

	if g.paused {
		// Check for 'C' to continue
		if ebiten.IsKeyPressed(ebiten.KeyC) {
			if !g.keyState[ebiten.KeyC] {
				g.paused = false
			}
			g.keyState[ebiten.KeyC] = true
		} else {
			g.keyState[ebiten.KeyC] = false
		}

		// Check for 'Q' to quit
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			g.terminated = true
			g.keyState[ebiten.KeyQ] = true
		} else {
			g.keyState[ebiten.KeyQ] = false
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Set the background color
	if g.hitCorner {
		screen.Fill(color.RGBA{0, 255, 0, 255}) // Flash green if hit a corner
	} else {
		screen.Fill(color.RGBA{0, 0, 255, 255}) // Default blue background
	}

	// Draw the logo
	op := &ebiten.DrawImageOptions{}
	scale := logoWidth / float64(g.logoImage.Bounds().Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(g.logoX, g.logoY)
	screen.DrawImage(g.logoImage, op)

	// Update window title with corner hits and elapsed time
	g.updateWindowTitle()

	if g.paused {
		g.drawPauseMenu(screen)
	}
}

func (g *Game) updateWindowTitle() {
	elapsedTime := time.Since(g.startTime)
	hours := int(elapsedTime.Hours())
	minutes := int(elapsedTime.Minutes()) % 60
	seconds := int(elapsedTime.Seconds()) % 60
	milliseconds := int(elapsedTime.Milliseconds()) % 1000
	title := fmt.Sprintf("Hits: %d | Time: %02d:%02d:%02d.%02d", g.cornerHits, hours, minutes, seconds, milliseconds/10)
	ebiten.SetWindowTitle(title)
}

func (g *Game) drawPauseMenu(screen *ebiten.Image) {
	// Draw the pause menu background
	pauseMenuWidth := 300
	pauseMenuHeight := 200
	pauseMenuX := (screenWidth - pauseMenuWidth) / 2
	pauseMenuY := (screenHeight - pauseMenuHeight) / 2
	ebitenutil.DrawRect(screen, float64(pauseMenuX), float64(pauseMenuY), float64(pauseMenuWidth), float64(pauseMenuHeight), color.RGBA{0, 0, 128, 255}) // Dark blue background

	// Draw the pause menu border
	borderThickness := 2.0
	ebitenutil.DrawRect(screen, float64(pauseMenuX), float64(pauseMenuY), float64(pauseMenuWidth), borderThickness, color.White)
	ebitenutil.DrawRect(screen, float64(pauseMenuX), float64(pauseMenuY), borderThickness, float64(pauseMenuHeight), color.White)
	ebitenutil.DrawRect(screen, float64(pauseMenuX), float64(pauseMenuY+pauseMenuHeight)-borderThickness, float64(pauseMenuWidth), borderThickness, color.White)
	ebitenutil.DrawRect(screen, float64(pauseMenuX+pauseMenuWidth)-borderThickness, float64(pauseMenuY), borderThickness, float64(pauseMenuHeight), color.White)

	// Draw the pause menu text
	pauseText := "PAUSED"
	continueText := "[C]ontinue"
	quitText := "[Q]uit"
	text.Draw(screen, pauseText, basicfont.Face7x13, pauseMenuX+pauseMenuWidth/2-len(pauseText)*7/2, pauseMenuY+50, color.White)
	text.Draw(screen, continueText, basicfont.Face7x13, pauseMenuX+pauseMenuWidth/2-len(continueText)*7/2, pauseMenuY+100, color.White)
	text.Draw(screen, quitText, basicfont.Face7x13, pauseMenuX+pauseMenuWidth/2-len(quitText)*7/2, pauseMenuY+150, color.White)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("DVD Logo Bouncer")

	logoImage, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(logoImageData))
	if err != nil {
		log.Fatal(err)
	}

	scale := logoWidth / float64(logoImage.Bounds().Dx())
	logoHeight := scale * float64(logoImage.Bounds().Dy())

	game := &Game{
		logoX:      float64(rand.Intn(screenWidth - int(logoWidth))),
		logoY:      float64(rand.Intn(screenHeight - int(logoHeight))),
		velocityX:  logoStartVelocity,
		velocityY:  logoStartVelocity,
		startTime:  time.Now(),
		logoImage:  logoImage,
		logoHeight: logoHeight,
		keyState:   make(map[ebiten.Key]bool),
	}

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
