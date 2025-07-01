package main

import (
	"rpg_demo/aseprite"

	"github.com/hajimehoshi/ebiten/v2"
)

// Orc represents an enemy orc character
type Orc struct {
	// Sprite and animation data
	sprite       *ebiten.Image
	asepriteFile *aseprite.File

	// Position and movement
	positionX  float64
	positionY  float64
	facingLeft bool

	// Animation state
	currentFrame  int
	frameTimer    float64
	frameDuration float64 // in seconds

	// Animation frame ranges
	idleFrameStart     int
	idleFrameEnd       int
	walkFrameStart     int
	walkFrameEnd       int
	attack01FrameStart int
	attack01FrameEnd   int
	attack02FrameStart int
	attack02FrameEnd   int
	hurtFrameStart     int
	hurtFrameEnd       int
	deathFrameStart    int
	deathFrameEnd      int

	// State management
	state OrcState

	// AI and movement
	walkSpeed   float64
	patrolLeft  float64 // Left boundary of patrol area
	patrolRight float64 // Right boundary of patrol area
	movingRight bool    // Direction of movement

	// Combat
	health     int
	maxHealth  int
	hurtTimer  float64 // Timer for hurt state duration
	knockbackX float64 // Knockback velocity

	// Death sequence
	deathTimer   float64 // Timer for death sequence
	flashTimer   float64 // Timer for flashing effect
	flashVisible bool    // Whether sprite is visible during flash
	flashCount   int     // Number of flashes completed
	shouldRemove bool    // Whether the orc should be removed
}

// OrcState represents the current state of the orc
type OrcState int

const (
	OrcStateIdle OrcState = iota
	OrcStateWalk
	OrcStateAttack01
	OrcStateAttack02
	OrcStateHurt
	OrcStateDeath
)

// NewOrc creates a new Orc instance
func NewOrc(x, y float64) (*Orc, error) {
	// Load the Orc Aseprite file
	aseFile, err := aseprite.LoadFile("assets/Orc.aseprite")
	if err != nil {
		return nil, err
	}

	// Get the first frame as an image
	frameImg, err := aseFile.GetFrameImage(0)
	if err != nil {
		return nil, err
	}

	orc := &Orc{
		sprite:        ebiten.NewImageFromImage(frameImg),
		asepriteFile:  aseFile,
		positionX:     x,
		positionY:     y,
		facingLeft:    false,
		currentFrame:  0,
		frameDuration: 0.1, // 100ms = 0.1 seconds
		frameTimer:    0,
		state:         OrcStateWalk, // Start walking
		walkSpeed:     2.0,          // Slower than player
		patrolLeft:    x - 150,      // Patrol 150 pixels left of starting position
		patrolRight:   x + 150,      // Patrol 150 pixels right of starting position
		movingRight:   true,         // Start moving right
		health:        3,            // Takes 3 hits to defeat
		maxHealth:     3,
		hurtTimer:     0,
		knockbackX:    0,
		deathTimer:    0,
		flashTimer:    0,
		flashVisible:  true,
		flashCount:    0,
		shouldRemove:  false,
	}

	// Initialize animation frame ranges from tags
	orc.initializeAnimationRanges()

	return orc, nil
}

// initializeAnimationRanges sets up the frame ranges for each animation
func (o *Orc) initializeAnimationRanges() {
	for _, tag := range o.asepriteFile.Tags {
		switch tag.Name {
		case "idle":
			o.idleFrameStart = int(tag.FromFrame)
			o.idleFrameEnd = int(tag.ToFrame)
		case "walk":
			o.walkFrameStart = int(tag.FromFrame)
			o.walkFrameEnd = int(tag.ToFrame)
		case "attack01":
			o.attack01FrameStart = int(tag.FromFrame)
			o.attack01FrameEnd = int(tag.ToFrame)
		case "attack02":
			o.attack02FrameStart = int(tag.FromFrame)
			o.attack02FrameEnd = int(tag.ToFrame)
		case "hurt":
			o.hurtFrameStart = int(tag.FromFrame)
			o.hurtFrameEnd = int(tag.ToFrame)
		case "Death":
			o.deathFrameStart = int(tag.FromFrame)
			o.deathFrameEnd = int(tag.ToFrame)
		}
	}

	// Start with walk animation since we begin walking
	o.currentFrame = o.walkFrameStart
}

// Update handles the orc's logic updates
func (o *Orc) Update(playerX float64) error {
	// Handle hurt state timing
	if o.state == OrcStateHurt {
		o.hurtTimer -= 1.0 / 60.0 // Decrease timer
		if o.hurtTimer <= 0 {
			// Hurt state finished, return to walking
			o.setState(OrcStateWalk)
		}
	}

	// Handle death sequence
	if o.state == OrcStateDeath {
		o.deathTimer -= 1.0 / 60.0 // Decrease timer
		if o.deathTimer <= 0 {
			// Start flashing sequence
			o.flashTimer -= 1.0 / 60.0
			if o.flashTimer <= 0 {
				// Toggle visibility
				o.flashVisible = !o.flashVisible
				o.flashTimer = 0.1 // Flash every 0.1 seconds

				if !o.flashVisible {
					o.flashCount++
				}

				// After 6 flashes (3 on/off cycles), mark for removal
				if o.flashCount >= 6 {
					o.shouldRemove = true
				}
			}
		}
	}

	// Handle knockback physics
	if o.knockbackX != 0 {
		o.positionX += o.knockbackX
		// Apply friction to knockback
		o.knockbackX *= 0.9
		// Stop knockback when it's very small
		if o.knockbackX > -1 && o.knockbackX < 1 {
			o.knockbackX = 0
		}
	}

	// Handle player-chasing AI (only when walking)
	if o.state == OrcStateWalk {
		// Move towards the player
		if playerX > o.positionX {
			// Player is to the right, move right
			o.positionX += o.walkSpeed
			o.facingLeft = false
		} else if playerX < o.positionX {
			// Player is to the left, move left
			o.positionX -= o.walkSpeed
			o.facingLeft = true
		}
		// If playerX == o.positionX, don't move horizontally
	}

	// Update animation timer
	o.frameTimer += 1.0 / 60.0 // Assuming 60 FPS

	// Check if it's time to advance to the next frame
	if o.frameTimer >= o.frameDuration {
		o.frameTimer = 0
		o.currentFrame++

		// Handle animation looping based on current state
		switch o.state {
		case OrcStateIdle:
			if o.currentFrame > o.idleFrameEnd {
				o.currentFrame = o.idleFrameStart
			}
		case OrcStateWalk:
			if o.currentFrame > o.walkFrameEnd {
				o.currentFrame = o.walkFrameStart
			}
		case OrcStateAttack01:
			if o.currentFrame > o.attack01FrameEnd {
				o.setState(OrcStateIdle)
			}
		case OrcStateAttack02:
			if o.currentFrame > o.attack02FrameEnd {
				o.setState(OrcStateIdle)
			}
		case OrcStateHurt:
			if o.currentFrame > o.hurtFrameEnd {
				// Don't change state here - let the timer handle it
				o.currentFrame = o.hurtFrameStart
			}
		case OrcStateDeath:
			if o.currentFrame > o.deathFrameEnd {
				// Stay on the last frame of death animation
				o.currentFrame = o.deathFrameEnd
			}
		}

		// Update the sprite image to the current frame
		if o.asepriteFile != nil {
			frameImg, err := o.asepriteFile.GetFrameImage(o.currentFrame)
			if err == nil {
				o.sprite = ebiten.NewImageFromImage(frameImg)
			}
		}
	}

	return nil
}

// setState changes the orc's state and resets animation
func (o *Orc) setState(newState OrcState) {
	if o.state == newState {
		return
	}

	o.state = newState
	o.frameTimer = 0

	// Set the starting frame for the new state
	switch newState {
	case OrcStateIdle:
		o.currentFrame = o.idleFrameStart
	case OrcStateWalk:
		o.currentFrame = o.walkFrameStart
	case OrcStateAttack01:
		o.currentFrame = o.attack01FrameStart
	case OrcStateAttack02:
		o.currentFrame = o.attack02FrameStart
	case OrcStateHurt:
		o.currentFrame = o.hurtFrameStart
	case OrcStateDeath:
		o.currentFrame = o.deathFrameStart
	}
}

// Draw renders the orc to the screen
func (o *Orc) Draw(screen *ebiten.Image) {
	if o.sprite == nil {
		return
	}

	// Don't draw if flashing and currently invisible
	if o.state == OrcStateDeath && o.deathTimer <= 0 && !o.flashVisible {
		return
	}

	opts := &ebiten.DrawImageOptions{}

	// Scale the sprite 10x larger (same as player)
	const scale = 10.0
	opts.GeoM.Scale(scale, scale)

	// Calculate sprite dimensions
	spriteWidth := float64(o.sprite.Bounds().Dx()) * scale
	spriteHeight := float64(o.sprite.Bounds().Dy()) * scale

	// Calculate final position
	finalX := (float64(screenWidth)-spriteWidth)/2 + o.positionX
	finalY := (float64(screenHeight)-spriteHeight)/2 + o.positionY

	// If facing left, flip around the center of the sprite
	if o.facingLeft {
		// Translate to center, flip, then translate back
		opts.GeoM.Translate(-spriteWidth/2, -spriteHeight/2)
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(spriteWidth/2, spriteHeight/2)
	}

	// Position the sprite at its final location
	opts.GeoM.Translate(finalX, finalY)

	screen.DrawImage(o.sprite, opts)
}

// GetBounds returns the collision bounds of the orc (adjusted for actual character size)
func (o *Orc) GetBounds() (x, y, width, height float64) {
	const scale = 10.0
	spriteWidth := float64(o.sprite.Bounds().Dx()) * scale
	spriteHeight := float64(o.sprite.Bounds().Dy()) * scale

	// Smaller collision box - only the core body area (8x8 pixels scaled up)
	// This makes it harder for the orc to hit the player
	charWidth := 8.0 * scale  // Smaller character width (scaled)
	charHeight := 8.0 * scale // Smaller character height (scaled)

	// Center the collision box within the sprite bounds
	finalX := (float64(screenWidth)-spriteWidth)/2 + o.positionX + (spriteWidth-charWidth)/2
	finalY := (float64(screenHeight)-spriteHeight)/2 + o.positionY + (spriteHeight-charHeight)/2

	return finalX, finalY, charWidth, charHeight
}

// CheckCollisionWithPlayer checks if the orc collides with the player (for damage to player)
func (o *Orc) CheckCollisionWithPlayer(playerX, playerY float64) bool {
	// Get orc bounds (already adjusted for character size)
	orcX, orcY, orcW, orcH := o.GetBounds()

	// Calculate player bounds with accurate character size
	const scale = 10.0
	spriteW := 100.0 * scale // Full sprite width
	spriteH := 100.0 * scale // Full sprite height

	// Player character collision box - smaller for more precise collision (8x8 pixels scaled up)
	// This matches the orc's collision box size for consistency
	playerCharW := 8.0 * scale // Smaller character width (scaled)
	playerCharH := 8.0 * scale // Smaller character height (scaled)

	// Calculate player sprite position (same as in main.go)
	playerSpriteX := (float64(screenWidth)-spriteW)/2 + playerX
	playerSpriteY := (float64(screenHeight)-spriteH)/2 + float64(screenHeight)*0.2

	// Center the collision box within the player sprite bounds
	playerFinalX := playerSpriteX + (spriteW-playerCharW)/2
	playerFinalY := playerSpriteY + (spriteH-playerCharH)/2

	// Simple AABB collision detection
	return playerFinalX < orcX+orcW &&
		playerFinalX+playerCharW > orcX &&
		playerFinalY < orcY+orcH &&
		playerFinalY+playerCharH > orcY
}

// CheckCollisionWithPlayerAttack checks if the orc is within the player's attack range and direction
func (o *Orc) CheckCollisionWithPlayerAttack(playerX, playerY float64, facingLeft bool) bool {
	// Get orc bounds (already adjusted for character size)
	orcX, orcY, orcW, orcH := o.GetBounds()

	// Calculate player bounds with directional attack range
	const scale = 10.0
	spriteW := 100.0 * scale // Full sprite width
	spriteH := 100.0 * scale // Full sprite height

	// Player attack range - larger than collision box (15x15 pixels scaled up)
	// This allows the player to hit the orc from a safer distance
	attackRangeW := 15.0 * scale // Larger attack width (scaled)
	attackRangeH := 15.0 * scale // Larger attack height (scaled)

	// Calculate player sprite position (same as in main.go)
	playerSpriteX := (float64(screenWidth)-spriteW)/2 + playerX
	playerSpriteY := (float64(screenHeight)-spriteH)/2 + float64(screenHeight)*0.2

	// Position attack range based on facing direction
	var playerAttackX, playerAttackY float64
	if facingLeft {
		// Attack range is to the left of the player
		playerAttackX = playerSpriteX + (spriteW-attackRangeW)/2 - attackRangeW/2
	} else {
		// Attack range is to the right of the player
		playerAttackX = playerSpriteX + (spriteW-attackRangeW)/2 + attackRangeW/2
	}
	playerAttackY = playerSpriteY + (spriteH-attackRangeH)/2

	// Simple AABB collision detection for directional attack range
	return playerAttackX < orcX+orcW &&
		playerAttackX+attackRangeW > orcX &&
		playerAttackY < orcY+orcH &&
		playerAttackY+attackRangeH > orcY
}

// TakeDamage handles the orc taking damage from player attacks
func (o *Orc) TakeDamage(attackerX float64) {
	// Don't take damage if already hurt or dead
	if o.state == OrcStateHurt || o.state == OrcStateDeath {
		return
	}

	o.health--

	if o.health <= 0 {
		// Orc dies
		o.setState(OrcStateDeath)
		o.deathTimer = 3.0 // Wait 3 seconds before flashing
	} else {
		// Orc gets hurt
		o.setState(OrcStateHurt)
		o.hurtTimer = 0.5 // Hurt state lasts 0.5 seconds

		// Apply knockback away from attacker
		if attackerX < o.positionX {
			// Attacker is to the left, knock orc right
			o.knockbackX = 30
		} else {
			// Attacker is to the right, knock orc left
			o.knockbackX = -30
		}
	}
}

// IsAlive returns whether the orc is still alive
func (o *Orc) IsAlive() bool {
	return o.state != OrcStateDeath
}

// ShouldRemove returns whether the orc should be removed from the game
func (o *Orc) ShouldRemove() bool {
	return o.shouldRemove
}

// GetHealth returns the current health of the orc
func (o *Orc) GetHealth() int {
	return o.health
}
