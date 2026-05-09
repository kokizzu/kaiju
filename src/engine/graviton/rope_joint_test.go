/******************************************************************************/
/* rope_joint_test.go                                                         */
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
	"testing"

	"kaijuengine.com/matrix"
)

func TestRopeJointAllowsSlack(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8

	a := addJointBody(&system, matrix.Vec3{-0.5, 0, 0}, RigidBodyTypeDynamic)
	b := addJointBody(&system, matrix.Vec3{0.5, 0, 0}, RigidBodyTypeDynamic)
	a.MotionState.LinearVelocity = matrix.Vec3Left().Scale(10)
	b.MotionState.LinearVelocity = matrix.Vec3Right().Scale(10)
	rope := system.NewRopeJoint(a, b, matrix.Vec3Zero(), matrix.Vec3Zero())
	rope.SetMaxLength(3)

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 1.0/60.0)

	if !rope.IsSlack() {
		t.Fatalf("expected rope to remain slack below max length, got length %f", rope.CurrentLength())
	}
	if rope.AccumulatedImpulse != 0 {
		t.Fatalf("expected slack rope to apply no impulse, got %f", rope.AccumulatedImpulse)
	}
	if !matrix.Vec3ApproxTo(a.MotionState.LinearVelocity, matrix.Vec3Left().Scale(10), 0.0001) ||
		!matrix.Vec3ApproxTo(b.MotionState.LinearVelocity, matrix.Vec3Right().Scale(10), 0.0001) {
		t.Fatalf("expected slack rope to leave velocities unchanged, got %v and %v",
			a.MotionState.LinearVelocity, b.MotionState.LinearVelocity)
	}
}

func TestRopeJointClampsMaxDistance(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	system.ConstraintPositionIterations = 8

	a := addJointBody(&system, matrix.Vec3{-2, 0, 0}, RigidBodyTypeDynamic)
	b := addJointBody(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	system.NewRopeJoint(a, b, matrix.Vec3Zero(), matrix.Vec3Zero()).SetMaxLength(2)

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	for range 120 {
		system.Step(workGroup, threads, 1.0/60.0)
	}

	distance := jointBodyDistance(a, b)
	if distance > 2.03 {
		t.Fatalf("expected rope to clamp max length 2 within tolerance, got %f", distance)
	}
}

func TestRopeJointDynamicToWorldAnchor(t *testing.T) {
	system := System{}
	system.Initialize()
	system.ConstraintPositionIterations = 8

	body := addJointBody(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeDynamic)
	system.NewRopeJointToWorld(body, matrix.Vec3Zero(), matrix.Vec3Zero()).SetMaxLength(2)

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	for range 240 {
		system.Step(workGroup, threads, 1.0/60.0)
	}

	position := body.Transform.WorldPosition()
	distance := position.Distance(matrix.Vec3Zero())
	if distance > 2.05 {
		t.Fatalf("expected dynamic-to-world rope to stay within max length 2, got %f at %v",
			distance, position)
	}
	if position.Y() >= -0.1 {
		t.Fatalf("expected anchored body to swing downward under gravity, got %v", position)
	}
}
