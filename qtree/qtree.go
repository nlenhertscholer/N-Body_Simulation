package qtree

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gen2brain/raylib-go/raymath"
	"proj3/phys"
)

const theta = 0.8    // Theta value to determine level of accuracy
const maxDepth = 800 // Helps avoid a stack overflow due to recurrsion
// - Lower if FPS starts getting too low; Make higher if more accuracy is wanted

// Barnes-Hut Tree (BHTree) is a QuadTree data structure
// that is used to approximate forces acting on each other during N-Body simulations
type BHTree struct {
	boundary rl.Rectangle // Boundary of the BHTree
	body     phys.Body    // Holds the bodies that belong in this BHTree
	divided  bool         // Whether this BHTree has subdivided or not
	nw       *BHTree      // Top Left of quadrant
	ne       *BHTree      // Top right of quadrant
	sw       *BHTree      // Bottom Left of quadrant
	se       *BHTree      // Bottom Right of quadrant
}

/*
 * Return a new BHTree
 */
func NewBHTree(bound rl.Rectangle) *BHTree {
	return &BHTree{bound, phys.Body{}, false,
		nil, nil, nil, nil}
}

/*
 * Insert a new body into the BHTree
 *
 * depth: Integer representing the maximum recursive depth
 */
func (q *BHTree) Insert(body phys.Body, depth int) {

	depth++

	if q.body.Mass == 0 {
		// A body has not been added here
		// Becomes an external node
		q.body = body

	} else if q.divided {
		// This is an internal node
		// Update this body's COM
		q.body = phys.AddBody(q.body, body)

		// Add this down the tree
		// I would put this in a different function but I want to limit the recursion
		if q.nw.contains(&body) {
			// Top left
			q.nw.Insert(body, depth)
		} else if q.ne.contains(&body) {
			// Top right
			q.ne.Insert(body, depth)
		} else if q.sw.contains(&body) {
			// Bottom left
			q.sw.Insert(body, depth)
		} else {
			// Bottom right
			q.se.Insert(body, depth)
		}

	} else {
		// This is an external node, create a center of Mass and subdivide
		otherBody := q.body
		q.body = phys.AddBody(otherBody, body)

		if depth < maxDepth {
			q.subdivide()

			// Insert the two bodies
			if q.nw.contains(&body) {
				// Top left
				q.nw.Insert(body, depth)
			} else if q.ne.contains(&body) {
				// Top right
				q.ne.Insert(body, depth)
			} else if q.sw.contains(&body) {
				// Bottom left
				q.sw.Insert(body, depth)
			} else {
				// Bottom right
				q.se.Insert(body, depth)
			}

			if q.nw.contains(&otherBody) {
				// Top left
				q.nw.Insert(otherBody, depth)
			} else if q.ne.contains(&otherBody) {
				// Top right
				q.ne.Insert(otherBody, depth)
			} else if q.sw.contains(&otherBody) {
				// Bottom left
				q.sw.Insert(otherBody, depth)
			} else {
				// Bottom right
				q.se.Insert(otherBody, depth)
			}

		}
	}
}

/*
 * Calculate the forces on the body from the entire tree
 */
func (q *BHTree) CalculateForces(body *phys.Body) {

	if body.Mass == 0 || q.body.Id == body.Id {
		// No need to calculate force
		return
	}

	if !q.divided {
		// External node - calculate full force
		body.AddForce(&q.body)
		return
	}

	// Get parameters to determine whether this body is sufficiently far away
	s := q.boundary.Width
	d := raymath.Vector2Distance(q.body.Position, body.Position)

	if s/d < theta {
		// This node is sufficiently far away to approximate using COM
		body.AddForce(&q.body)
	} else {
		// Not sufficiently far away - calculate for each body
		q.nw.CalculateForces(body)
		q.ne.CalculateForces(body)
		q.sw.CalculateForces(body)
		q.se.CalculateForces(body)
	}
}

/*
 * Draw the tree to the screen in GUI mode
 */
func (q *BHTree) DrawTree() {

	rl.DrawRectangleLinesEx(q.boundary, 1, rl.Blue)

	if q.divided {
		q.nw.DrawTree()
		q.ne.DrawTree()
		q.sw.DrawTree()
		q.se.DrawTree()
	}
}

/*
 * Subdivide the BHTree into smaller sections
 */
func (q *BHTree) subdivide() {

	// Create the new boundaries for the different sections
	nw := rl.NewRectangle(q.boundary.X, q.boundary.Y,
		q.boundary.Width/2, q.boundary.Height/2)
	ne := rl.NewRectangle(q.boundary.X+q.boundary.Width/2, q.boundary.Y,
		q.boundary.Width/2, q.boundary.Height/2)
	sw := rl.NewRectangle(q.boundary.X, q.boundary.Y+q.boundary.Height/2,
		q.boundary.Width/2, q.boundary.Height/2)
	se := rl.NewRectangle(q.boundary.X+q.boundary.Width/2, q.boundary.Y+q.boundary.Height/2,
		q.boundary.Width/2, q.boundary.Height/2)

	// Create the new BHTrees
	q.nw = NewBHTree(nw)
	q.ne = NewBHTree(ne)
	q.se = NewBHTree(sw)
	q.sw = NewBHTree(se)

	q.divided = true
}

/*
 * Contains checks to see if a body is in the correct rectangle
 */
func (q *BHTree) contains(body *phys.Body) bool {
	return body.Position.X >= q.boundary.X &&
		body.Position.X <= q.boundary.X+q.boundary.Width &&
		body.Position.Y >= q.boundary.Y &&
		body.Position.Y <= q.boundary.Y+q.boundary.Height
}
