/******************************************************************************/
/* hinge_joint.go                                                             */
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

// HingeJoint keeps two anchors coincident and aligns each body's hinge axis.
// Relative rotation is constrained around the two axes perpendicular to the
// hinge, leaving rotation around the hinge axis free.
type HingeJoint struct {
	BodyA        *RigidBody
	BodyB        *RigidBody
	LocalAnchorA matrix.Vec3
	LocalAnchorB matrix.Vec3
	LocalAxisA   matrix.Vec3
	LocalAxisB   matrix.Vec3

	Stiffness                matrix.Float
	BiasFactor               matrix.Float
	PositionCorrectionFactor matrix.Float
	Slop                     matrix.Float
	MaxCorrection            matrix.Float
	WarmStarting             bool

	AccumulatedAnchorImpulse  matrix.Vec3
	AccumulatedAngularImpulse matrix.Vec2

	constraint  *Constraint
	anchorRows  [3]ConstraintSolverRow
	angularRows [2]AngularConstraintSolverRow
}

func NewHingeJoint(bodyA, bodyB *RigidBody, localAnchorA, localAnchorB, localAxisA, localAxisB matrix.Vec3) *HingeJoint {
	return &HingeJoint{
		BodyA:                    bodyA,
		BodyB:                    bodyB,
		LocalAnchorA:             localAnchorA,
		LocalAnchorB:             localAnchorB,
		LocalAxisA:               safeNormal(localAxisA, matrix.Vec3Right()),
		LocalAxisB:               safeNormal(localAxisB, matrix.Vec3Right()),
		Stiffness:                defaultDistanceJointStiffness,
		BiasFactor:               defaultDistanceJointBiasFactor,
		PositionCorrectionFactor: defaultDistanceJointPositionCorrectionFactor,
		Slop:                     defaultDistanceJointSlop,
		MaxCorrection:            defaultDistanceJointMaxCorrection,
	}
}

func NewHingeJointAtWorldAnchor(bodyA, bodyB *RigidBody, worldAnchor, worldAxis matrix.Vec3) *HingeJoint {
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	return NewHingeJoint(
		bodyA,
		bodyB,
		LocalAnchor(bodyA, worldAnchor),
		LocalAnchor(bodyB, worldAnchor),
		LocalAxis(bodyA, axis),
		LocalAxis(bodyB, axis),
	)
}

func NewHingeJointToWorld(body *RigidBody, localAnchor, worldAnchor, localAxis, worldAxis matrix.Vec3) *HingeJoint {
	return NewHingeJoint(
		body,
		nil,
		localAnchor,
		worldAnchor,
		localAxis,
		safeNormal(worldAxis, matrix.Vec3Right()),
	)
}

func (j *HingeJoint) WorldAnchorA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyA, j.LocalAnchorA)
}

func (j *HingeJoint) WorldAnchorB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return WorldAnchor(j.BodyB, j.LocalAnchorB)
}

func (j *HingeJoint) WorldAxisA() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Right()
	}
	return WorldAxis(j.BodyA, j.LocalAxisA)
}

func (j *HingeJoint) WorldAxisB() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Right()
	}
	return WorldAxis(j.BodyB, j.LocalAxisB)
}

func (j *HingeJoint) CurrentAnchorError() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return j.WorldAnchorB().Subtract(j.WorldAnchorA())
}

func (j *HingeJoint) CurrentAngularError() matrix.Vec3 {
	if j == nil {
		return matrix.Vec3Zero()
	}
	return hingeAngularError(j.WorldAxisA(), j.WorldAxisB())
}

func (j *HingeJoint) SetWorldAnchors(worldAnchorA, worldAnchorB matrix.Vec3) {
	if j == nil {
		return
	}
	j.LocalAnchorA = LocalAnchor(j.BodyA, worldAnchorA)
	j.LocalAnchorB = LocalAnchor(j.BodyB, worldAnchorB)
	j.AccumulatedAnchorImpulse = matrix.Vec3Zero()
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) SetWorldAxis(worldAxis matrix.Vec3) {
	if j == nil {
		return
	}
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	j.LocalAxisA = LocalAxis(j.BodyA, axis)
	j.LocalAxisB = LocalAxis(j.BodyB, axis)
	j.AccumulatedAngularImpulse = matrix.Vec2{}
	WakeConstrainedBodies(j.BodyA, j.BodyB)
}

func (j *HingeJoint) IsStretched() bool {
	if j == nil {
		return false
	}
	return j.CurrentAnchorError().Length() > j.slop() ||
		j.CurrentAngularError().Length() > j.slop()
}

func (j *HingeJoint) prepare(deltaTime matrix.Float) {
	if j == nil {
		return
	}
	j.prepareAnchorRows(deltaTime)
	j.prepareAngularRows(deltaTime)
}

func (j *HingeJoint) prepareAnchorRows(deltaTime matrix.Float) {
	anchorA := j.WorldAnchorA()
	anchorB := j.WorldAnchorB()
	error := anchorB.Subtract(anchorA)
	for i, axis := range pointJointAxes {
		row := &j.anchorRows[i]
		row.SetWorldAnchors(j.BodyA, j.BodyB, anchorA, anchorB, axis)
		row.EffectiveMass *= j.stiffness()
		row.Bias = j.bias(error.Dot(axis), deltaTime)
		row.AccumulatedImpulse = 0
		if j.WarmStarting {
			row.AccumulatedImpulse = j.AccumulatedAnchorImpulse[i]
			row.ApplyImpulse(row.AccumulatedImpulse)
		}
	}
}

func (j *HingeJoint) prepareAngularRows(deltaTime matrix.Float) {
	axisA := j.WorldAxisA()
	axisB := j.WorldAxisB()
	hingeAxis := safeNormal(axisA.Add(axisB), axisA)
	error := hingeAngularError(axisA, axisB)
	axes := hingeConstraintAxes(hingeAxis)
	for i, axis := range axes {
		row := &j.angularRows[i]
		row.SetWorldAxis(j.BodyA, j.BodyB, axis)
		row.EffectiveMass *= j.stiffness()
		row.Bias = j.bias(error.Dot(axis), deltaTime)
		row.AccumulatedImpulse = 0
		if j.WarmStarting {
			row.AccumulatedImpulse = j.AccumulatedAngularImpulse[i]
			row.ApplyImpulse(row.AccumulatedImpulse)
		}
	}
}

func (j *HingeJoint) solveVelocity() {
	if j == nil {
		return
	}
	for i := range j.anchorRows {
		j.anchorRows[i].Solve()
		if j.WarmStarting {
			j.AccumulatedAnchorImpulse[i] = j.anchorRows[i].AccumulatedImpulse
		}
	}
	for i := range j.angularRows {
		j.angularRows[i].Solve()
		if j.WarmStarting {
			j.AccumulatedAngularImpulse[i] = j.angularRows[i].AccumulatedImpulse
		}
	}
}

func (j *HingeJoint) solvePosition() {
	if j == nil {
		return
	}
	j.solveAnchorPosition()
	j.solveAngularPosition()
}

func (j *HingeJoint) solveAnchorPosition() {
	error := j.CurrentAnchorError()
	if error.Length() <= j.slop() {
		return
	}
	invMassA := j.BodyA.inverseMass()
	invMassB := j.BodyB.inverseMass()
	invMassSum := invMassA + invMassB
	if invMassSum <= contactEpsilon {
		return
	}
	correction := j.clampedCorrection(error)
	correction = correction.Scale(1.0 / invMassSum)
	moveBody(j.BodyA, correction.Scale(invMassA))
	moveBody(j.BodyB, correction.Scale(-invMassB))
}

func (j *HingeJoint) solveAngularPosition() {
	error := j.CurrentAngularError()
	if error.Length() <= j.slop() {
		return
	}
	axis := safeNormal(error, matrix.Vec3Right())
	invA := AngularAxisEffectiveMass(j.BodyA, axis)
	invB := AngularAxisEffectiveMass(j.BodyB, axis)
	invSum := invA + invB
	if invSum <= contactEpsilon {
		return
	}
	correction := j.clampedAngularCorrection(error)
	rotateBody(j.BodyA, correction.Scale(invA/invSum))
	rotateBody(j.BodyB, correction.Scale(-invB/invSum))
}

func (j *HingeJoint) bias(error, deltaTime matrix.Float) matrix.Float {
	if deltaTime <= 0 {
		deltaTime = defaultDistanceJointTimeStep
	}
	if matrix.Abs(error) <= j.slop() {
		return 0
	}
	return error * j.biasFactor() / deltaTime
}

func (j *HingeJoint) clampedCorrection(error matrix.Vec3) matrix.Vec3 {
	correction := error.Scale(j.positionCorrectionFactor() * j.stiffness())
	maxCorrection := j.maxCorrection()
	length := correction.Length()
	if length > maxCorrection && length > matrix.FloatSmallestNonzero {
		correction = correction.Scale(maxCorrection / length)
	}
	return correction
}

func (j *HingeJoint) clampedAngularCorrection(error matrix.Vec3) matrix.Vec3 {
	correction := error.Scale(j.positionCorrectionFactor() * j.stiffness())
	maxCorrection := j.maxCorrection()
	length := correction.Length()
	if length > maxCorrection && length > matrix.FloatSmallestNonzero {
		correction = correction.Scale(maxCorrection / length)
	}
	return correction
}

func (j *HingeJoint) stiffness() matrix.Float {
	if j.Stiffness < 0 {
		return 0
	}
	return matrix.Clamp(j.Stiffness, 0, 1)
}

func (j *HingeJoint) biasFactor() matrix.Float {
	if j.BiasFactor < 0 {
		return 0
	}
	return j.BiasFactor
}

func (j *HingeJoint) positionCorrectionFactor() matrix.Float {
	if j.PositionCorrectionFactor < 0 {
		return 0
	}
	return j.PositionCorrectionFactor
}

func (j *HingeJoint) slop() matrix.Float {
	if j.Slop <= 0 {
		return defaultDistanceJointSlop
	}
	return j.Slop
}

func (j *HingeJoint) maxCorrection() matrix.Float {
	if j.MaxCorrection <= 0 {
		return defaultDistanceJointMaxCorrection
	}
	return j.MaxCorrection
}

func LocalAxis(body *RigidBody, worldAxis matrix.Vec3) matrix.Vec3 {
	axis := safeNormal(worldAxis, matrix.Vec3Right())
	if body == nil {
		return axis
	}
	rotation := body.Rotation()
	rotation.Inverse()
	return safeNormal(rotation.Rotate(axis), matrix.Vec3Right())
}

func WorldAxis(body *RigidBody, localAxis matrix.Vec3) matrix.Vec3 {
	axis := safeNormal(localAxis, matrix.Vec3Right())
	if body == nil {
		return axis
	}
	return safeNormal(body.Rotation().Rotate(axis), matrix.Vec3Right())
}

func hingeConstraintAxes(hingeAxis matrix.Vec3) [2]matrix.Vec3 {
	axis := safeNormal(hingeAxis, matrix.Vec3Right())
	first := safeNormal(axis.Orthogonal(), matrix.Vec3Up())
	second := safeNormal(axis.Cross(first), matrix.Vec3Forward())
	return [2]matrix.Vec3{first, second}
}

func hingeAngularError(axisA, axisB matrix.Vec3) matrix.Vec3 {
	a := safeNormal(axisA, matrix.Vec3Right())
	b := safeNormal(axisB, matrix.Vec3Right())
	cross := a.Cross(b)
	sin := cross.Length()
	dot := matrix.Clamp(a.Dot(b), -1, 1)
	if sin <= contactEpsilon {
		if dot >= 0 {
			return matrix.Vec3Zero()
		}
		return safeNormal(a.Orthogonal(), matrix.Vec3Up()).Scale(matrix.Atan2(0, dot))
	}
	return cross.Scale(matrix.Atan2(sin, dot) / sin)
}

func rotateBody(body *RigidBody, angularCorrection matrix.Vec3) {
	if body == nil || body.inverseInertia().IsZero() {
		return
	}
	angle := angularCorrection.Length()
	if angle <= contactEpsilon {
		return
	}
	delta := matrix.QuaternionAxisAngle(angularCorrection.Scale(1.0/angle), angle)
	current := matrix.QuaternionFromEuler(body.Transform.Rotation())
	next := delta.Multiply(current)
	next.Normalize()
	body.Transform.SetRotation(next.ToEuler())
}
