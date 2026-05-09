/******************************************************************************/
/* rope_joint.go                                                              */
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

import "kaijuengine.com/matrix"

// RopeJoint keeps two anchors from separating beyond MaxLength. Unlike a
// DistanceJoint, it is inactive while the anchors are within the limit.
type RopeJoint struct {
	BodyA        *RigidBody
	BodyB        *RigidBody
	LocalAnchorA matrix.Vec3
	LocalAnchorB matrix.Vec3
	MaxLength    matrix.Float

	Stiffness                matrix.Float
	BiasFactor               matrix.Float
	PositionCorrectionFactor matrix.Float
	Slop                     matrix.Float
	MaxCorrection            matrix.Float
	WarmStarting             bool

	AccumulatedImpulse matrix.Float

	constraint *Constraint
	row        ConstraintSolverRow
	lastAxis   matrix.Vec3
	taut       bool
}

func NewRopeJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB matrix.Vec3) *RopeJoint {
	joint := &RopeJoint{
		BodyA:                    bodyA,
		BodyB:                    bodyB,
		LocalAnchorA:             localAnchorA,
		LocalAnchorB:             localAnchorB,
		Stiffness:                defaultDistanceJointStiffness,
		BiasFactor:               defaultDistanceJointBiasFactor,
		PositionCorrectionFactor: defaultDistanceJointPositionCorrectionFactor,
		Slop:                     defaultDistanceJointSlop,
		MaxCorrection:            defaultDistanceJointMaxCorrection,
		lastAxis:                 matrix.Vec3Right(),
	}
	joint.MaxLength = joint.CurrentLength()
	return joint
}

func NewRopeJointAtWorldAnchors(bodyA, bodyB *RigidBody, worldAnchorA, worldAnchorB matrix.Vec3) *RopeJoint {
	return NewRopeJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchorA),
		LocalAnchor(bodyB, worldAnchorB),
	)
}

func NewRopeJointToWorld(body *RigidBody, localAnchor, worldAnchor matrix.Vec3) *RopeJoint {
	return NewRopeJoint(body, nil, localAnchor, worldAnchor)
}

func (j *RopeJoint) WorldAnchorA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyA, j.LocalAnchorA)
}

func (j *RopeJoint) WorldAnchorB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyB, j.LocalAnchorB)
}

func (j *RopeJoint) CurrentLength() matrix.Float {
	if j == nil {
		return 0
	}
	return j.WorldAnchorB().Subtract(j.WorldAnchorA()).Length()
}

func (j *RopeJoint) SetWorldAnchors(worldAnchorA, worldAnchorB matrix.Vec3) {
	if j == nil {
		return
	}
	j.LocalAnchorA = LocalAnchor(j.BodyA, worldAnchorA)
	j.LocalAnchorB = LocalAnchor(j.BodyB, worldAnchorB)
	j.MaxLength = j.CurrentLength()
	j.AccumulatedImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *RopeJoint) SetMaxLength(maxLength matrix.Float) {
	if j == nil {
		return
	}
	j.MaxLength = matrix.Max(maxLength, 0)
	j.AccumulatedImpulse = 0
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *RopeJoint) IsStretched() bool {
	return j != nil && j.CurrentLength() > j.MaxLength+j.slop()
}

func (j *RopeJoint) IsSlack() bool {
	return j != nil && !j.IsStretched()
}

func (j *RopeJoint) prepare(deltaTime matrix.Float) {
	if j == nil {
		return
	}
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	axis, length, ok := j.axisAndLength(anchorA, anchorB)
	j.taut = ok && length > j.MaxLength
	if !j.taut {
		j.row = ConstraintSolverRow{}
		j.AccumulatedImpulse = 0
		return
	}
	j.row.SetWorldAnchors(j.BodyA, j.BodyB, anchorA, anchorB, axis)
	j.row.EffectiveMass *= j.stiffness()
	j.row.Bias = j.bias(length, deltaTime)
	j.row.SetImpulseLimits(-matrix.Inf(1), 0)
	j.row.AccumulatedImpulse = 0
	if j.WarmStarting {
		j.row.AccumulatedImpulse = matrix.Min(j.AccumulatedImpulse, 0)
		j.row.ApplyImpulse(j.row.AccumulatedImpulse)
	}
}

func (j *RopeJoint) solveVelocity() {
	if j == nil || !j.taut {
		return
	}
	j.row.Solve()
	j.AccumulatedImpulse = j.row.AccumulatedImpulse
}

func (j *RopeJoint) solvePosition() {
	if j == nil {
		return
	}
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	axis, length, ok := j.axisAndLength(anchorA, anchorB)
	if !ok || length <= j.MaxLength+j.slop() {
		return
	}
	invMassA := j.BodyA.inverseMass()
	invMassB := j.BodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	correction := (length - j.MaxLength) * j.positionCorrectionFactor() * j.stiffness()
	correction = matrix.Min(correction, j.maxCorrection())
	correction /= invMassSum
	moveBody(j.BodyA, axis.Scale(correction*invMassA))
	moveBody(j.BodyB, axis.Scale(-correction*invMassB))
}

func (j *RopeJoint) axisAndLength(anchorA, anchorB matrix.Vec3) (matrix.Vec3, matrix.Float, bool) {
	delta := anchorB.Subtract(anchorA)
	lengthSq := delta.LengthSquared()
	if lengthSq <= distanceJointMinLength*distanceJointMinLength {
		axis := j.lastAxis
		if axis.LengthSquared() <= distanceJointMinLength*distanceJointMinLength {
			axis = matrix.Vec3Right()
		}
		j.lastAxis = axis.Normal()
		return j.lastAxis, 0, false
	}
	length := matrix.Sqrt(lengthSq)
	j.lastAxis = delta.Scale(1.0 / length)
	return j.lastAxis, length, true
}

func (j *RopeJoint) bias(length, deltaTime matrix.Float) matrix.Float {
	if deltaTime <= 0 {
		deltaTime = defaultDistanceJointTimeStep
	}
	return matrix.Max(length-j.MaxLength, 0) * j.biasFactor() / deltaTime
}

func (j *RopeJoint) stiffness() matrix.Float {
	if j.Stiffness < 0 {
		return 0
	}
	return matrix.Clamp(j.Stiffness, 0, 1)
}

func (j *RopeJoint) biasFactor() matrix.Float {
	if j.BiasFactor < 0 {
		return 0
	}
	return j.BiasFactor
}

func (j *RopeJoint) positionCorrectionFactor() matrix.Float {
	if j.PositionCorrectionFactor < 0 {
		return 0
	}
	return j.PositionCorrectionFactor
}

func (j *RopeJoint) slop() matrix.Float {
	if j.Slop <= 0 {
		return defaultDistanceJointSlop
	}
	return j.Slop
}

func (j *RopeJoint) maxCorrection() matrix.Float {
	if j.MaxCorrection <= 0 {
		return defaultDistanceJointMaxCorrection
	}
	return j.MaxCorrection
}
