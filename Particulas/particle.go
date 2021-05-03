package main

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
)

type Particles struct {
	particles                   []Particle
	points                      []float32
	color                       mgl32.Vec3
	position, velocity          mgl32.Vec3
	minLife, maxLife, amplitude float32
}

type Particle struct {
	x, y, z, alpha *float32
	life0, life    float32
	velocity       mgl32.Vec3
}

func (p *Particles) addParticle(index int, life float32) {
	pos := index * 7
	p.points[pos], p.points[pos+1], p.points[pos+2] = p.position.X(), p.position.Y(), p.position.Z()
	p.points[pos+3], p.points[pos+4], p.points[pos+5], p.points[pos+6] = p.color.X(), p.color.Y(), p.color.Z(), 1

	p.particles[index] = Particle{
		x:        &p.points[pos],
		y:        &p.points[pos+1],
		z:        &p.points[pos+2],
		life0:    life,
		life:     life,
		velocity: p.velocity,
		alpha:    &p.points[pos+6],
	}
}

func NewParticles(numParticles int, color mgl32.Vec3, position, velocity mgl32.Vec3, minLife, maxLife, amplitude float32) *Particles {
	particles := Particles{
		particles: make([]Particle, numParticles),
		points:    make([]float32, numParticles*7),
		color:     color,
		position:  position,
		velocity:  velocity,
		minLife:   minLife,
		maxLife:   maxLife,
		amplitude: amplitude,
	}
	for i := 0; i < numParticles; i++ {
		life := minLife + rand.Float32()*(maxLife-minLife)
		particles.addParticle(i, life)
	}
	return &particles
}

func (p *Particles) Update(dT float32, position, velocity mgl32.Vec3) {
	p.position = position
	for i := range p.particles {
		particle := &p.particles[i]

		particle.life -= dT
		if particle.life <= 0 {
			*particle.x = p.position.X()
			*particle.y = p.position.Y()
			*particle.z = p.position.Z()

			particle.velocity = p.velocity.Add(mgl32.Vec3{rand.Float32()*2*p.amplitude - p.amplitude, rand.Float32()*2*p.amplitude - p.amplitude, rand.Float32()*2*p.amplitude - p.amplitude})

			life := p.minLife + rand.Float32()*(p.maxLife-p.minLife)
			particle.life = life
		} else {
			*particle.x += dT * particle.velocity.X()
			*particle.y += dT * particle.velocity.Y()
			*particle.z += dT * particle.velocity.Z()
			*particle.alpha = particle.life / particle.life0
		}
	}
}
