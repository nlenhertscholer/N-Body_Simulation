package phys

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gen2brain/raylib-go/raymath"
	"math"
)

// Constants
const G = 1              // Gravitational force - Not being realistic
const MaxDistance = 2500 // Added so numbers don't blow up
const dt = 0.4           // Timestep to multiply force and velocity with
const RadiusCoeff = 3

// Body is a physics body which can be affected by gravity
type Body struct {
	Mass     float32
	Position rl.Vector2
	Velocity rl.Vector2
	Radius   float32
	Id       int
	Force    rl.Vector2 // Force to be applied each timestep
}

/*
 * Return a new body with the specified inputs
 */
func NewBody(mass float32, id int, pos, vel rl.Vector2) Body {
	radius := mass * RadiusCoeff
	return Body{mass, pos, vel,
		radius, id, rl.NewVector2(0, 0)}
}

/*
 * Adds the force due to the other body
 */
func (b *Body) AddForce(other *Body) {

	distance := raymath.Vector2Distance(b.Position, other.Position)

	// Clamp the distance between the larger object's radius and MaxDistance
	distance = raymath.Clamp(distance, float32(math.Max(float64(b.Radius),
		float64(other.Radius))), MaxDistance)

	// Unit vector pointing in the direction of the applied force
	force := raymath.Vector2Subtract(other.Position, b.Position)
	raymath.Vector2Divide(&force, distance)

	// Calculate the strength of the force
	strength := G * other.Mass / (distance * distance)

	// Calculate the new force
	raymath.Vector2Scale(&force, strength)

	// Add the force to this object
	b.Force = raymath.Vector2Add(b.Force, force)

}

/*
 * Zero out the force
 */
func (b *Body) ZeroForce() {
	b.Force = rl.NewVector2(0, 0)
}

/*
 * Apply the force to update this object's position
 * Uses Euler-Cromer method
 */
func (b *Body) Update() rl.Vector2 {

	// Multiply force by the timestep
	raymath.Vector2Scale(&b.Force, dt)

	// Update the velocity
	b.Velocity = raymath.Vector2Add(b.Velocity, b.Force)
	vel := b.Velocity              // Copy to scale velocity
	raymath.Vector2Scale(&vel, dt) // Scale it by timestep

	// Update the position
	b.Position = raymath.Vector2Add(b.Position, vel)

	return b.Position

}

/*
 * Combine the bodies and return a body representing their Center of Mass (COM)
 * and combined mass and velocity
 */
func AddBody(b1 Body, b2 Body) Body {

	m := b1.Mass + b2.Mass // Combined mass

	// X and Y of COM
	x := (b1.Position.X*b1.Mass + b2.Position.X*b2.Mass) / m
	y := (b1.Position.Y*b1.Mass + b2.Position.Y*b2.Mass) / m

	// X and Y of VCM (velocity of center of mass)
	dx := (b1.Velocity.X*b1.Mass + b2.Velocity.X*b2.Mass) / m
	dy := (b1.Velocity.Y*b1.Mass + b2.Velocity.Y*b2.Mass) / m

	return NewBody(m, -1, rl.NewVector2(x, y), rl.NewVector2(dx, dy))
}
