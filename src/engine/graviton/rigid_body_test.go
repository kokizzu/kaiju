/******************************************************************************/
/* rigid_body_test.go                                                         */
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

func TestRigidBodyApplyForceChangesLinearVelocityOnStep(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.SetMass(2, matrix.Vec3One())
	body.ApplyForce(matrix.Vec3{4, 0, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.5)
	expected := matrix.Vec3{1, 0, 0}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
	if !body.MotionState.Acceleration.IsZero() {
		t.Fatalf("expected force acceleration accumulator to reset, got %v", body.MotionState.Acceleration)
	}
}

func TestRigidBodyApplyForceAtPointChangesAngularVelocityOnStep(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())
	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.SetMass(2, matrix.Vec3One())
	body.ApplyForceAtPoint(matrix.Vec3{2, 0, 0}, matrix.Vec3{0, 1, 0})
	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()
	system.Step(workGroup, threads, 0.5)
	expectedLinear := matrix.Vec3{0.5, 0, 0}
	expectedAngular := matrix.Vec3{0, 0, -1}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expectedLinear, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expectedLinear, body.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(body.MotionState.AngularVelocity, expectedAngular, 0.0001) {
		t.Fatalf("expected angular velocity %v, got %v", expectedAngular, body.MotionState.AngularVelocity)
	}
	if !body.MotionState.AngularAcceleration.IsZero() {
		t.Fatalf("expected torque acceleration accumulator to reset, got %v", body.MotionState.AngularAcceleration)
	}
}

func TestRigidBodyApplyImpulseChangesLinearVelocityImmediately(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.ApplyImpulse(matrix.Vec3{4, 0, 0})
	expected := matrix.Vec3{2, 0, 0}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
	if !body.MotionState.AngularVelocity.IsZero() {
		t.Fatalf("expected central impulse not to change angular velocity, got %v", body.MotionState.AngularVelocity)
	}
}

func TestRigidBodyApplyImpulseAtPointChangesAngularVelocityImmediately(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.ApplyImpulseAtPoint(matrix.Vec3{2, 0, 0}, matrix.Vec3{0, 1, 0})
	expectedLinear := matrix.Vec3{1, 0, 0}
	expectedAngular := matrix.Vec3{0, 0, -2}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expectedLinear, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expectedLinear, body.MotionState.LinearVelocity)
	}
	if !matrix.Vec3ApproxTo(body.MotionState.AngularVelocity, expectedAngular, 0.0001) {
		t.Fatalf("expected angular velocity %v, got %v", expectedAngular, body.MotionState.AngularVelocity)
	}
}

func TestRigidBodyApplyImpulseWakesSleepingBody(t *testing.T) {
	body := testRigidBody(Shape{}, matrix.Vec3Zero())
	body.SetMass(2, matrix.Vec3One())
	body.Sleep()
	body.ApplyImpulse(matrix.Vec3{4, 0, 0})
	expected := matrix.Vec3{2, 0, 0}
	if body.Simulation.IsSleeping {
		t.Fatal("expected impulse to wake body")
	}
	if !matrix.Vec3ApproxTo(body.MotionState.LinearVelocity, expected, 0.0001) {
		t.Fatalf("expected linear velocity %v, got %v", expected, body.MotionState.LinearVelocity)
	}
}
