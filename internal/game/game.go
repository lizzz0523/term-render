package game

import (
	"math"

	"term-render/internal/geo"
	mdl "term-render/internal/model"
	"term-render/internal/renderer"

	"github.com/gdamore/tcell/v2"
)

const EyeHeight = 1.5

var lightDir = geo.NewVec3(0.3, 0.5, -0.8).Norm()

type Player struct {
	Pos   geo.Vec3
	Angle float64
}

type Enemy struct {
	Pos       geo.Vec3
	Alive     bool
	DeathTime float64
}

type Game struct {
	Player     Player
	Enemies    []Enemy
	GunModel   *mdl.Model
	EnemyModel *mdl.Model
	GameTime   float64
}

func NewGame(gunModel *mdl.Model, enemyModel *mdl.Model) *Game {
	return &Game{
		Player: Player{Pos: geo.NewVec3(0, EyeHeight, 0)},
		Enemies: []Enemy{
			{Pos: geo.NewVec3(-3, 0, -5), Alive: true},
			{Pos: geo.NewVec3(3, 0, -5), Alive: true},
			{Pos: geo.NewVec3(0, 0, -6), Alive: true},
			{Pos: geo.NewVec3(-2, 0, -8), Alive: true},
			{Pos: geo.NewVec3(2, 0, -8), Alive: true},
		},
		GunModel:   gunModel,
		EnemyModel: enemyModel,
	}
}

func (g *Game) Intersect(ro, rd geo.Vec3) (renderer.Hit, bool) {
	bestT := math.MaxFloat64
	var bestHit renderer.Hit
	found := false

	if h, ok := g.intersectGun(ro, rd); ok {
		t := h.Point.Sub(ro).Len()
		if t < bestT {
			bestT, bestHit, found = t, h, true
		}
	}

	for i := range g.Enemies {
		if !g.Enemies[i].Alive && g.GameTime-g.Enemies[i].DeathTime > 0.3 {
			continue
		}
		if h, ok := g.intersectEnemy(ro, rd, i); ok {
			t := h.Point.Sub(ro).Len()
			if t < bestT {
				if !g.Enemies[i].Alive {
					h.Normal = lightDir
				}
				bestT, bestHit, found = t, h, true
			}
		}
	}

	if h, ok := g.intersectFloor(ro, rd); ok {
		t := h.Point.Sub(ro).Len()
		if t < bestT {
			bestT, bestHit, found = t, h, true
		}
	}

	return bestHit, found
}

func (g *Game) intersectGun(ro, rd geo.Vec3) (renderer.Hit, bool) {
	offset := geo.NewVec3(0.5, -0.35, 1.5).RotY(g.Player.Angle)
	gunPos := ro.Add(offset)
	localRo := ro.Sub(gunPos).RotY(-g.Player.Angle)
	localRd := rd.RotY(-g.Player.Angle)
	hit, ok := g.GunModel.Intersect(localRo, localRd)
	if !ok {
		return renderer.Hit{}, false
	}
	return renderer.Hit{
		Point:  hit.Point.RotY(g.Player.Angle).Add(gunPos),
		Normal: hit.Normal.RotY(g.Player.Angle),
	}, true
}

func (g *Game) intersectEnemy(ro, rd geo.Vec3, i int) (renderer.Hit, bool) {
	enemyPos := g.Enemies[i].Pos.Add(geo.NewVec3(0, 0.9, 0))
	localRo := ro.Sub(enemyPos)
	localRd := rd
	hit, ok := g.EnemyModel.Intersect(localRo, localRd)
	if !ok {
		return renderer.Hit{}, false
	}
	return renderer.Hit{Point: hit.Point.Add(enemyPos), Normal: hit.Normal}, true
}

func (g *Game) intersectFloor(ro, rd geo.Vec3) (renderer.Hit, bool) {
	if rd.Y >= 0 {
		return renderer.Hit{}, false
	}
	t := -ro.Y / rd.Y
	if t < 0 {
		return renderer.Hit{}, false
	}
	return renderer.Hit{Point: ro.Add(rd.Mul(t)), Normal: geo.NewVec3(0, 1, 0)}, true
}

func (g *Game) HandleInput(ev tcell.Event) bool {
	switch e := ev.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEscape:
			return false
		case tcell.KeyLeft:
			g.Player.Pos.X -= math.Cos(g.Player.Angle) * 0.3
			g.Player.Pos.Z += math.Sin(g.Player.Angle) * 0.3
		case tcell.KeyRight:
			g.Player.Pos.X += math.Cos(g.Player.Angle) * 0.3
			g.Player.Pos.Z -= math.Sin(g.Player.Angle) * 0.3
		case tcell.KeyUp:
			g.Player.Pos.X += math.Sin(g.Player.Angle) * 0.3
			g.Player.Pos.Z += math.Cos(g.Player.Angle) * 0.3
		case tcell.KeyDown:
			g.Player.Pos.X -= math.Sin(g.Player.Angle) * 0.3
			g.Player.Pos.Z -= math.Cos(g.Player.Angle) * 0.3
		case tcell.KeyRune:
			switch e.Rune() {
			case 'q', 'Q':
				return false
			case 'z', 'Z':
				g.Player.Angle -= 0.1
			case 'x', 'X':
				g.Player.Angle += 0.1
			case ' ':
				g.shoot()
			}
		}
	}
	return true
}

func (g *Game) shoot() {
	rd := geo.NewVec3(0, 0, 1).Norm().RotY(g.Player.Angle)
	ro := g.Player.Pos
	for i := range g.Enemies {
		if !g.Enemies[i].Alive {
			continue
		}
		if _, ok := g.intersectEnemy(ro, rd, i); ok {
			g.Enemies[i].Alive = false
			g.Enemies[i].DeathTime = g.GameTime
			break
		}
	}
}
