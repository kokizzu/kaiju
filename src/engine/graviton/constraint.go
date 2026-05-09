/******************************************************************************/
/* constraint.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package graviton

import "kaijuengine.com/engine/pooling"

type ConstraintType uint8

const (
	ConstraintTypeUnknown ConstraintType = iota
	ConstraintTypeGeneric
	ConstraintTypeDistance
)

// Constraint stores the lifecycle and endpoints for a future Graviton
// constraint solver. BodyA and BodyB form a body-body constraint; either body
// may be nil to represent a body-world constraint.
type Constraint struct {
	Type     ConstraintType
	BodyA    *RigidBody
	BodyB    *RigidBody
	Rows     []ConstraintSolverRow
	Distance *DistanceJoint
	Active   bool
	Enabled  bool
	poolId   pooling.PoolGroupId
	id       pooling.PoolIndex
	pooled   bool
}

func (c *Constraint) IsBodyBody() bool {
	return c != nil && c.BodyA != nil && c.BodyB != nil
}

func (c *Constraint) IsBodyWorld() bool {
	return c != nil && ((c.BodyA != nil && c.BodyB == nil) ||
		(c.BodyA == nil && c.BodyB != nil))
}

func (c *Constraint) BodiesValid() bool {
	if c == nil {
		return false
	}
	if c.BodyA == nil && c.BodyB == nil {
		return false
	}
	if c.BodyA != nil && !constraintBodyValid(c.BodyA) {
		return false
	}
	if c.BodyB != nil && !constraintBodyValid(c.BodyB) {
		return false
	}
	return true
}

func (c *Constraint) IsValid() bool {
	return c != nil && c.pooled && c.Active && c.Enabled && c.BodiesValid()
}

func (c *Constraint) SetBodies(bodyA, bodyB *RigidBody) {
	c.BodyA = bodyA
	c.BodyB = bodyB
	if c.Distance != nil {
		c.Distance.BodyA = bodyA
		c.Distance.BodyB = bodyB
	}
	c.disableIfBodiesInvalid()
}

func (c *Constraint) disableIfBodiesInvalid() {
	if !c.BodiesValid() {
		c.Active = false
		c.Enabled = false
	}
}

func (c *Constraint) detachBody(body *RigidBody) {
	if body == nil {
		return
	}
	if c.BodyA == body {
		c.BodyA = nil
	}
	if c.BodyB == body {
		c.BodyB = nil
	}
	if c.Distance != nil {
		c.Distance.BodyA = c.BodyA
		c.Distance.BodyB = c.BodyB
	}
	c.Active = false
	c.Enabled = false
}

func constraintBodyValid(body *RigidBody) bool {
	return body != nil && body.pooled
}
