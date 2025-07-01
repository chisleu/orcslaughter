package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"

	"rpg_demo/aseprite"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 1536
	screenHeight = 1024
)

// PlayerState represents the current state of the player
type PlayerState int

const (
	PlayerStateAlive PlayerState = iota
	PlayerStateHurt
	PlayerStateDying
	PlayerStateDead
)

// Game represents our game state
type Game struct {
	soldierSprite   *ebiten.Image
	backgroundImage *ebiten.Image
	asepriteFile    *aseprite.File

	// Animation state
	currentFrame     int
	frameTimer       float64
	frameDuration    float64 // in seconds
	idleFrameStart   int
	idleFrameEnd     int
	walkFrameStart   int
	walkFrameEnd     int
	attackFrameStart int
	attackFrameEnd   int
	hurtFrameStart   int
	hurtFrameEnd     int

	// Movement and sprite state
	positionX   float64
	isWalking   bool
	facingLeft  bool
	walkSpeed   float64
	isAttacking bool

	// Player state
	playerState     PlayerState
	playerHealth    float64
	deathFrameStart int
	deathFrameEnd   int
	deathTimer      float64 // Timer for death sequence
	flashTimer      float64 // Timer for flashing effect
	flashVisible    bool    // Whether player sprite is visible during flash
	flashCount      int     // Number of flashes completed

	// Audio
	audioContext *audio.Context
	musicPlayer  *audio.Player
	attackPlayer *audio.Player
	orcHitPlayer *audio.Player
	orcDiePlayer *audio.Player

	// Enemies and scoring
	orcs          []*Orc  // Multiple orcs
	orcPrevHealth int     // Track previous orc health to detect damage
	orcsKilled    int     // Counter for killed orcs
	spawnTimer    float64 // Timer for spawning new orcs
	spawnInterval float64 // Time between spawns (decreases as game progresses)
}

// Update handles game logic updates
func (g *Game) Update() error {
	g.handlePlayerInput()
	g.updatePlayerAnimation()
	g.updatePlayerDeath()
	g.updateOrcLogic()
	return nil
}

// Draw handles rendering
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw background first
	if g.backgroundImage != nil {
		screen.DrawImage(g.backgroundImage, &ebiten.DrawImageOptions{})
	}

	// Draw the soldier sprite (don't draw if flashing and currently invisible)
	if g.soldierSprite != nil && !(g.playerState == PlayerStateDying && g.deathTimer <= 0 && !g.flashVisible) {
		opts := &ebiten.DrawImageOptions{}

		// Scale the sprite 10x larger
		const scale = 10.0
		opts.GeoM.Scale(scale, scale)

		// Calculate sprite dimensions
		spriteWidth := float64(g.soldierSprite.Bounds().Dx()) * scale
		spriteHeight := float64(g.soldierSprite.Bounds().Dy()) * scale

		// Calculate final position
		finalX := (float64(screenWidth)-spriteWidth)/2 + g.positionX
		finalY := (float64(screenHeight)-spriteHeight)/2 + float64(screenHeight)*0.2

		// If facing left, flip around the center of the sprite
		if g.facingLeft {
			// Translate to center, flip, then translate back
			opts.GeoM.Translate(-spriteWidth/2, -spriteHeight/2)
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(spriteWidth/2, spriteHeight/2)
		}

		// Position the sprite at its final location
		opts.GeoM.Translate(finalX, finalY)

		screen.DrawImage(g.soldierSprite, opts)
	}

	// Draw all orcs
	for _, orc := range g.orcs {
		if orc != nil {
			orc.Draw(screen)
		}
	}

	// Draw kill counter in top-left corner
	killText := fmt.Sprintf("Orcs Killed: %d", g.orcsKilled)
	text.Draw(screen, killText, basicfont.Face7x13, 20, 30, color.RGBA{255, 255, 255, 255})

	// Draw lifebar at bottom-center of screen
	barWidth := 300.0
	barHeight := 20.0
	barX := (float64(screenWidth) - barWidth) / 2
	barY := float64(screenHeight) - 60 // 60 pixels from bottom

	// Draw background (dark red)
	backgroundBar := ebiten.NewImage(int(barWidth), int(barHeight))
	backgroundBar.Fill(color.RGBA{100, 0, 0, 255}) // Dark red
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(barX, barY)
	screen.DrawImage(backgroundBar, opts)

	// Draw health bar (green, proportional to health)
	healthPercent := g.playerHealth / 100.0
	if healthPercent < 0 {
		healthPercent = 0
	}
	healthWidth := barWidth * healthPercent
	if healthWidth > 0 {
		healthBar := ebiten.NewImage(int(healthWidth), int(barHeight))
		healthBar.Fill(color.RGBA{0, 255, 0, 255}) // Bright green
		healthOpts := &ebiten.DrawImageOptions{}
		healthOpts.GeoM.Translate(barX, barY)
		screen.DrawImage(healthBar, healthOpts)
	}

	// Draw health text
	healthText := fmt.Sprintf("Health: %.0f%%", g.playerHealth)
	text.Draw(screen, healthText, basicfont.Face7x13, int(barX), int(barY-10), color.RGBA{255, 255, 255, 255})
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// loadImageFromFile loads an image from a file and converts it to an Ebiten image
func loadImageFromFile(filename string) (*ebiten.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("RPG Demo - Aseprite Loading")

	game := &Game{}

	// Initialize audio context
	game.audioContext = audio.NewContext(44100)

	// Load and play background music
	musicFile, err := os.Open("assets/soundtrack.mp3")
	if err != nil {
		log.Fatalf("Failed to load soundtrack.mp3: %v", err)
	}
	defer musicFile.Close()

	// Decode the MP3 file
	musicStream, err := mp3.DecodeWithoutResampling(musicFile)
	if err != nil {
		log.Fatalf("Failed to decode soundtrack.mp3: %v", err)
	}

	// Create an infinite loop stream
	loopStream := audio.NewInfiniteLoop(musicStream, musicStream.Length())

	// Create the music player
	game.musicPlayer, err = game.audioContext.NewPlayer(loopStream)
	if err != nil {
		log.Fatalf("Failed to create music player: %v", err)
	}

	// Set volume to low level (30% of maximum)
	game.musicPlayer.SetVolume(0.3)

	// Start playing the music
	game.musicPlayer.Play()

	// Load attack sound effect
	attackFile, err := os.Open("assets/attack.mp3")
	if err != nil {
		log.Fatalf("Failed to load attack.mp3: %v", err)
	}
	defer attackFile.Close()

	// Decode the attack sound MP3 file
	attackStream, err := mp3.DecodeWithoutResampling(attackFile)
	if err != nil {
		log.Fatalf("Failed to decode attack.mp3: %v", err)
	}

	// Create the attack sound player
	game.attackPlayer, err = game.audioContext.NewPlayer(attackStream)
	if err != nil {
		log.Fatalf("Failed to create attack sound player: %v", err)
	}

	// Set volume for attack sound (slightly higher than background music)
	game.attackPlayer.SetVolume(0.5)

	// Load orc hit sound effect
	orcHitFile, err := os.Open("assets/orc_hit.mp3")
	if err != nil {
		log.Fatalf("Failed to load orc_hit.mp3: %v", err)
	}
	defer orcHitFile.Close()

	// Decode the orc hit sound MP3 file
	orcHitStream, err := mp3.DecodeWithoutResampling(orcHitFile)
	if err != nil {
		log.Fatalf("Failed to decode orc_hit.mp3: %v", err)
	}

	// Create the orc hit sound player
	game.orcHitPlayer, err = game.audioContext.NewPlayer(orcHitStream)
	if err != nil {
		log.Fatalf("Failed to create orc hit sound player: %v", err)
	}

	// Set volume for orc hit sound
	game.orcHitPlayer.SetVolume(0.4)

	// Load orc die sound effect
	orcDieFile, err := os.Open("assets/orc_die.mp3")
	if err != nil {
		log.Fatalf("Failed to load orc_die.mp3: %v", err)
	}
	defer orcDieFile.Close()

	// Decode the orc die sound MP3 file
	orcDieStream, err := mp3.DecodeWithoutResampling(orcDieFile)
	if err != nil {
		log.Fatalf("Failed to decode orc_die.mp3: %v", err)
	}

	// Create the orc die sound player
	game.orcDiePlayer, err = game.audioContext.NewPlayer(orcDieStream)
	if err != nil {
		log.Fatalf("Failed to create orc die sound player: %v", err)
	}

	// Set volume for orc die sound
	game.orcDiePlayer.SetVolume(0.4)

	// Load the background image
	backgroundImg, err := loadImageFromFile("assets/background.png")
	if err != nil {
		log.Fatalf("Failed to load background.png: %v", err)
	}
	game.backgroundImage = backgroundImg

	// Load the Soldier Aseprite file
	aseFile, err := aseprite.LoadFile("assets/Soldier.aseprite")
	if err != nil {
		log.Fatalf("Failed to load Soldier.aseprite: %v", err)
	}
	game.asepriteFile = aseFile

	// Get the first frame as an image
	frameImg, err := aseFile.GetFrameImage(0)
	if err != nil {
		log.Fatalf("Failed to get frame image: %v", err)
	}

	// Convert to Ebiten image
	game.soldierSprite = ebiten.NewImageFromImage(frameImg)

	// Initialize animation state for "Idle", "Walk", "Attack01", "Hurt", and "Death" tags
	var idleTag, walkTag, attackTag, hurtTag, deathTag *aseprite.Tag
	for _, tag := range aseFile.Tags {
		if tag.Name == "Idle" {
			idleTag = tag
		} else if tag.Name == "Walk" {
			walkTag = tag
		} else if tag.Name == "Attack02" {
			attackTag = tag
		} else if tag.Name == "Hurt" {
			hurtTag = tag
		} else if tag.Name == "Death" {
			deathTag = tag
		}
	}

	if idleTag != nil {
		game.idleFrameStart = int(idleTag.FromFrame)
		game.idleFrameEnd = int(idleTag.ToFrame)
		log.Printf("Found Idle animation: frames %d-%d", game.idleFrameStart, game.idleFrameEnd)
	} else {
		log.Printf("Warning: Idle animation tag not found")
		game.idleFrameStart = 0
		game.idleFrameEnd = 0
	}

	if walkTag != nil {
		game.walkFrameStart = int(walkTag.FromFrame)
		game.walkFrameEnd = int(walkTag.ToFrame)
		log.Printf("Found Walk animation: frames %d-%d", game.walkFrameStart, game.walkFrameEnd)
	} else {
		log.Printf("Warning: Walk animation tag not found")
		game.walkFrameStart = 6
		game.walkFrameEnd = 13
	}

	if attackTag != nil {
		game.attackFrameStart = int(attackTag.FromFrame)
		game.attackFrameEnd = int(attackTag.ToFrame)
		log.Printf("Found Attack01 animation: frames %d-%d", game.attackFrameStart, game.attackFrameEnd)
	} else {
		log.Printf("Warning: Attack01 animation tag not found")
		game.attackFrameStart = 14
		game.attackFrameEnd = 19
	}

	if hurtTag != nil {
		game.hurtFrameStart = int(hurtTag.FromFrame)
		game.hurtFrameEnd = int(hurtTag.ToFrame)
		log.Printf("Found Hurt animation: frames %d-%d", game.hurtFrameStart, game.hurtFrameEnd)
	} else {
		log.Printf("Warning: Hurt animation tag not found")
		game.hurtFrameStart = 20
		game.hurtFrameEnd = 25
	}

	if deathTag != nil {
		game.deathFrameStart = int(deathTag.FromFrame)
		game.deathFrameEnd = int(deathTag.ToFrame)
		log.Printf("Found Death animation: frames %d-%d", game.deathFrameStart, game.deathFrameEnd)
	} else {
		log.Printf("Warning: Death animation tag not found")
		game.deathFrameStart = 26
		game.deathFrameEnd = 31
	}

	// Initialize movement and animation state
	game.currentFrame = game.idleFrameStart
	game.frameDuration = 0.1 // 100ms = 0.1 seconds
	game.frameTimer = 0
	game.positionX = 0
	game.isWalking = false
	game.facingLeft = false
	game.walkSpeed = 5.0
	game.isAttacking = false
	game.playerState = PlayerStateAlive
	game.playerHealth = 100.0 // Initialize player health to 100%
	game.deathTimer = 0
	game.flashTimer = 0
	game.flashVisible = true
	game.flashCount = 0
	game.orcsKilled = 0

	// Initialize spawn system
	game.orcs = make([]*Orc, 0)
	game.spawnTimer = 0
	game.spawnInterval = 9.0 // Start with 9 seconds between spawns (tripled)

	// Create the first orc enemy
	orc, err := NewOrc(300, float64(screenHeight)*0.2) // Position orc to the right of center
	if err != nil {
		log.Fatalf("Failed to create orc: %v", err)
	}
	game.orcs = append(game.orcs, orc)

	log.Printf("Loaded Aseprite file: %dx%d, %d frames, %d bpp",
		aseFile.Header.Width, aseFile.Header.Height,
		aseFile.Header.Frames, aseFile.Header.ColorDepth)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
