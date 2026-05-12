/******************************************************************************/
/* point_joint_test.go                                                        */
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

import (
	"math"
	"testing"

	"kaijuengine.com/matrix"
)

func TestPointJointKeepsAnchorsTogether(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	a := addJointBody(&system, matrix.Vec3{-2, 0.5, 0}, RigidBodyTypeDynamic)
	b := addJointBody(&system, matrix.Vec3{2, -0.25, 0.75}, RigidBodyTypeDynamic)
	b.Transform.SetRotation(matrix.Vec3{0, 45, 30})
	a.MotionState.LinearVelocity = matrix.Vec3{-6, 2, -1}
	b.MotionState.LinearVelocity = matrix.Vec3{5, -1, 3}
	joint := system.NewPointJoint(a, b, matrix.Vec3{0.5, 0.25, 0}, matrix.Vec3{-0.25, 0, 0.5})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	distance := joint.WorldAnchorA().Distance(joint.WorldAnchorB())
	if distance > 0.03 {
		t.Fatalf("expected point joint anchors to coincide, got distance %f at %v and %v",
			distance, joint.WorldAnchorA(), joint.WorldAnchorB())
	}
}

func TestPointJointAllowsRotation(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	body := addJointBody(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.MotionState.AngularVelocity = matrix.Vec3{0, matrix.Float(math.Pi), 0}
	joint := system.NewPointJointToWorld(body, matrix.Vec3Zero(), matrix.Vec3Zero())
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 60 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	if joint.WorldAnchorA().Distance(joint.WorldAnchorB()) > 0.001 {
		t.Fatalf("expected center anchor to remain fixed while rotating, got %v and %v",
			joint.WorldAnchorA(), joint.WorldAnchorB())
	}
	if !matrix.Vec3ApproxTo(body.MotionState.AngularVelocity, matrix.Vec3{0, matrix.Float(math.Pi), 0}, 0.0001) {
		t.Fatalf("expected point joint not to constrain angular velocity, got %v",
			body.MotionState.AngularVelocity)
	}
	if matrix.Vec3ApproxTo(body.Transform.Rotation(), matrix.Vec3Zero(), 0.001) {
		t.Fatalf("expected body rotation to remain free, got %v", body.Transform.Rotation())
	}
}

func TestPointJointBodyToWorld(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8
	body := addJointBody(&system, matrix.Vec3{2, 3, -1}, RigidBodyTypeDynamic)
	body.Transform.SetRotation(matrix.Vec3{0, 90, 0})
	body.MotionState.LinearVelocity = matrix.Vec3{4, -3, 2}
	joint := system.NewPointJointToWorld(body, matrix.Vec3{0.5, -0.25, 0.75}, matrix.Vec3Zero())
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}
	distance := joint.WorldAnchorA().Distance(matrix.Vec3Zero())
	if distance > 0.03 {
		t.Fatalf("expected body-world point joint anchor at world origin, got distance %f at %v",
			distance, joint.WorldAnchorA())
	}
}
