package main

import (
	"fmt"
	_ "image/png"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 5

	play     = 1
	menu     = 2
	gameOver = 3

	gravity      = 0.3
	maxVelocityY = 5

	chipmunkSize = 32

	totNumEnemies = 200
)

var (
	tiles     = make([]*Tile, 0, 0)
	runClouds = false
)

type Game struct {
	player       Player
	enemies      []*Enemy
	enemyBullets []*EnemyBullet
	gameMode     int

	layers        [][]int
	world         *ebiten.Image
	camera        Camera
	ominousClouds OminousClouds
	portal        Portal

	gameOver   bool
	playerWon  bool
	timeToExit time.Time
}

func init() {
	initAnimation()
	initBackgroundImg()
	initWorldImg()
}

func calcAliveEnemies(enemies []*Enemy) int {
	numAliveEnemies := 0
	for _, enemy := range enemies {
		if enemy.isAlive {
			numAliveEnemies += 1
		}
	}
	return numAliveEnemies
}

func (g *Game) Update() error {
	if (g.gameOver || g.playerWon) && time.Now().After(g.timeToExit) {
		os.Exit(0)
	}

	if ebiten.IsKeyPressed(ebiten.KeyO) {
		runClouds = true
		g.ominousClouds.StartClouds()
	}
	if ebiten.IsKeyPressed(ebiten.KeyP) {
		runClouds = false
		g.ominousClouds.StopClouds()
	}

	if ebiten.IsKeyPressed(ebiten.KeyY) {
		g.player.isResting = false
		g.player.isAttacking = true
	}

	g.camera.update(g)

	switch g.gameMode {
	case play:

		g.player.updatePlayer()
		g.UpdateEnemies()
		g.UpdateBullets()
		if g.isPlayerHit() {
			if g.player.health <= 0 {
				g.gameMode = gameOver
			}
		}

	case gameOver:
		fmt.Printf("Game Over! :(")
		g.player.health = 100
		g.gameMode = play

	default:
		g.gameMode = play
	}

	g.ominousClouds.UpdateClouds()

	aliveEnemies := calcAliveEnemies(g.enemies)
	if aliveEnemies < totNumEnemies {
		g.createEnemies(totNumEnemies - aliveEnemies)
	}

	if !g.playerWon && !g.gameOver {
		if g.player.y16 > 1700 {
			g.gameOver = true
			g.timeToExit = time.Now().Add(time.Second * time.Duration(5))
		}

		if g.player.health <= 0 {
			g.gameOver = true
			g.timeToExit = time.Now().Add(time.Second * time.Duration(5))
		}

		if g.player.x16 >= g.portal.x16 {
			g.playerWon = true
			g.timeToExit = time.Now().Add(time.Second * time.Duration(5))
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground()
	g.drawWorld()
	g.drawPortal()
	if g.player.invulnerable {
		nanoseconds := (time.Now().UnixNano() - g.player.lastInvulnerableStart.UnixNano())
		milliseconds := nanoseconds / 1000000
		if milliseconds%100 > 50 {
			g.drawCharacter()
		}
	} else {
		g.drawCharacter()
	}
	g.drawEnemies()
	g.DrawBullets()

	// sx, sy := frameOX+i*frameWidth, 0
	if g.gameMode == play {
		for _, tile := range tiles {
			tile.DrawTile(g.world)
		}
	}

	g.ominousClouds.DrawClouds(g.world)

	// Anything relative to world must be drawn on g.world before calling
	// Render()
	g.camera.Render(g.world, screen)
	numAliveEnemies := calcAliveEnemies(g.enemies)

	DrawOverlay(screen, g.player.health, numAliveEnemies, g)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
