package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// spawnOrc creates a new orc at a random off-screen position with increasing speed
func (g *Game) spawnOrc() {
	// Randomly choose left or right side of screen (50/50 chance)
	var spawnX float64
	if len(g.orcs)%2 == 0 {
		// Spawn on the left side (off-screen)
		spawnX = -float64(screenWidth)/2 - 200
	} else {
		// Spawn on the right side (off-screen)
		spawnX = float64(screenWidth)/2 + 200
	}

	// Create new orc with increased speed based on kills
	orc, err := NewOrc(spawnX, float64(screenHeight)*0.2)
	if err != nil {
		log.Printf("Failed to create new orc: %v", err)
		return
	}

	// Increase orc speed based on kills (each kill makes orcs 5% faster)
	speedMultiplier := 1.0 + (float64(g.orcsKilled) * 0.05)
	orc.walkSpeed = 2.0 * speedMultiplier

	// Add to orcs slice
	g.orcs = append(g.orcs, orc)

	// Decrease spawn interval slightly (make spawns faster)
	g.spawnInterval = g.spawnInterval * 0.95
	if g.spawnInterval < 0.5 {
		g.spawnInterval = 0.5 // Minimum spawn interval of 0.5 seconds
	}
}

// handlePlayerInput processes player input for movement and attacks
func (g *Game) handlePlayerInput() {
	// Handle attack input (only if not already attacking and not hurt)
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !g.isAttacking && g.playerState == PlayerStateAlive {
		g.isAttacking = true
		g.currentFrame = g.attackFrameStart
		g.frameTimer = 0

		// Play attack sound effect
		if g.attackPlayer != nil {
			g.attackPlayer.Rewind()
			g.attackPlayer.Play()
		}
	}

	// Handle movement input (only if not attacking and not hurt)
	if !g.isAttacking && g.playerState == PlayerStateAlive {
		wasWalking := g.isWalking
		g.isWalking = false

		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			g.isWalking = true
			g.facingLeft = true
			g.positionX -= g.walkSpeed
			// Keep within screen bounds
			if g.positionX < -float64(screenWidth)/2 {
				g.positionX = -float64(screenWidth) / 2
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			g.isWalking = true
			g.facingLeft = false
			g.positionX += g.walkSpeed
			// Keep within screen bounds
			if g.positionX > float64(screenWidth)/2 {
				g.positionX = float64(screenWidth) / 2
			}
		}

		// Switch animation if walking state changed
		if g.isWalking != wasWalking {
			if g.isWalking {
				// Switch to walk animation
				g.currentFrame = g.walkFrameStart
				g.frameTimer = 0
			} else {
				// Switch to idle animation
				g.currentFrame = g.idleFrameStart
				g.frameTimer = 0
			}
		}
	}
}

// updatePlayerDeath handles player death sequence and flashing
func (g *Game) updatePlayerDeath() {
	if g.playerState == PlayerStateDying {
		// Handle death sequence
		g.deathTimer -= 1.0 / 60.0 // Decrease timer
		if g.deathTimer <= 0 {
			// Start flashing sequence
			g.flashTimer -= 1.0 / 60.0
			if g.flashTimer <= 0 {
				// Toggle visibility
				g.flashVisible = !g.flashVisible
				g.flashTimer = 0.1 // Flash every 0.1 seconds

				if !g.flashVisible {
					g.flashCount++
				}

				// After 6 flashes (3 on/off cycles), exit the game
				if g.flashCount >= 6 {
					log.Printf("Game Over! Player died after killing %d orcs.", g.orcsKilled)
					os.Exit(0)
				}
			}
		}
	}
}

// updatePlayerAnimation handles player animation updates
func (g *Game) updatePlayerAnimation() {
	// Update animation timer
	g.frameTimer += 1.0 / 60.0 // Assuming 60 FPS

	// Check if it's time to advance to the next frame
	if g.frameTimer >= g.frameDuration {
		g.frameTimer = 0
		g.currentFrame++

		// Handle animation based on current state
		if g.playerState == PlayerStateDying {
			if g.currentFrame > g.deathFrameEnd {
				// Stay on the last frame of death animation
				g.currentFrame = g.deathFrameEnd
			}
		} else if g.playerState == PlayerStateHurt {
			if g.currentFrame > g.hurtFrameEnd {
				// Hurt animation finished, return to alive state
				g.playerState = PlayerStateAlive
				g.currentFrame = g.idleFrameStart
			}
		} else if g.isAttacking {
			if g.currentFrame > g.attackFrameEnd {
				// Attack animation finished, return to appropriate state
				g.isAttacking = false
				if g.isWalking {
					g.currentFrame = g.walkFrameStart
				} else {
					g.currentFrame = g.idleFrameStart
				}
			}
		} else if g.isWalking {
			if g.currentFrame > g.walkFrameEnd {
				g.currentFrame = g.walkFrameStart
			}
		} else {
			if g.currentFrame > g.idleFrameEnd {
				g.currentFrame = g.idleFrameStart
			}
		}

		// Update the sprite image to the current frame
		if g.asepriteFile != nil {
			frameImg, err := g.asepriteFile.GetFrameImage(g.currentFrame)
			if err == nil {
				g.soldierSprite = ebiten.NewImageFromImage(frameImg)
			}
		}
	}
}

// updateOrcLogic handles orc updates, interactions, and spawning
func (g *Game) updateOrcLogic() {
	// Update spawn timer
	g.spawnTimer += 1.0 / 60.0 // Assuming 60 FPS

	// Check if it's time to spawn a new orc
	if g.spawnTimer >= g.spawnInterval {
		g.spawnOrc()
		g.spawnTimer = 0 // Reset spawn timer
	}

	// Update all orcs and handle interactions
	for i := len(g.orcs) - 1; i >= 0; i-- {
		orc := g.orcs[i]
		if orc == nil {
			continue
		}

		// Store previous health to detect damage
		prevHealth := orc.GetHealth()
		wasAlive := orc.IsAlive()

		orc.Update(g.positionX)

		// Check if orc should be removed after death sequence
		if orc.ShouldRemove() {
			g.orcsKilled++ // Increment kill counter
			// Remove the orc from the slice
			g.orcs = append(g.orcs[:i], g.orcs[i+1:]...)
			continue
		}

		// Check if player attack hits this orc (using directional attack range)
		if g.isAttacking && orc.IsAlive() && orc.CheckCollisionWithPlayerAttack(g.positionX, 0, g.facingLeft) {
			// Player attack hits the orc
			orc.TakeDamage(g.positionX)

			// Check if orc took damage and play appropriate sound
			currentHealth := orc.GetHealth()
			if currentHealth < prevHealth {
				if currentHealth <= 0 && wasAlive {
					// Orc died - play death sound
					if g.orcDiePlayer != nil {
						g.orcDiePlayer.Rewind()
						g.orcDiePlayer.Play()
					}
				} else {
					// Orc took damage but didn't die - play hit sound
					if g.orcHitPlayer != nil {
						g.orcHitPlayer.Rewind()
						g.orcHitPlayer.Play()
					}
				}
			}
		}

		// Check for collision between player and this orc (only if orc is alive and player is not already hurt or dying)
		if orc.IsAlive() && g.playerState == PlayerStateAlive && orc.CheckCollisionWithPlayer(g.positionX, 0) {
			// Player takes damage
			g.playerHealth -= 10.0
			if g.playerHealth <= 0 {
				g.playerHealth = 0
				// Player dies - start death sequence
				g.playerState = PlayerStateDying
				g.currentFrame = g.deathFrameStart
				g.frameTimer = 0
				g.deathTimer = 3.0    // Wait 3 seconds before flashing
				g.isAttacking = false // Cancel any ongoing attack
				g.isWalking = false   // Cancel any ongoing movement
			} else {
				// Set player to hurt state and start hurt animation
				g.playerState = PlayerStateHurt
				g.currentFrame = g.hurtFrameStart
				g.frameTimer = 0
				g.isAttacking = false // Cancel any ongoing attack
				g.isWalking = false   // Cancel any ongoing movement
			}

			// Simple knockback effect - push player away from orc (5x stronger knockback)
			// Compare player position directly with orc position (both use same coordinate system)
			if g.positionX < orc.positionX {
				// Player is to the left of orc, push player further left (away from orc)
				g.positionX -= 100
			} else {
				// Player is to the right of orc, push player further right (away from orc)
				g.positionX += 100
			}

			// Keep player within screen bounds after knockback
			if g.positionX < -float64(screenWidth)/2 {
				g.positionX = -float64(screenWidth) / 2
			}
			if g.positionX > float64(screenWidth)/2 {
				g.positionX = float64(screenWidth) / 2
			}

			// Only take damage from one orc per frame
			break
		}
	}
}
