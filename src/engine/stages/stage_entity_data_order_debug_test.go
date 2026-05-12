/******************************************************************************/
/* stage_entity_data_order_debug_test.go                                      */
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

//go:build debug

/******************************************************************************/
/* stage_entity_data_order_debug_test.go                                      */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/

package stages

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
)

var debugStageOrderFailures []string

type debugStageOrderConstraintData struct {
	Target string
}

func (d debugStageOrderConstraintData) Init(e *engine.Entity, host *engine.Host) {
	target := host.EntityById(engine.EntityId(d.Target))
	if target == nil {
		debugStageOrderFailures = append(debugStageOrderFailures, "target entity was not registered")
		return
	}
	if _, ok := host.Physics().FindEntity(e); !ok {
		debugStageOrderFailures = append(debugStageOrderFailures, "source body was not staged")
		return
	}
	if _, ok := host.Physics().FindEntity(target); !ok {
		debugStageOrderFailures = append(debugStageOrderFailures, "target body was not staged")
		return
	}
	if host.Physics().AddDistanceJoint(e, target, matrix.Vec3Zero(), matrix.Vec3Zero()) == nil {
		debugStageOrderFailures = append(debugStageOrderFailures, "distance joint was not created")
	}
}

func (d debugStageOrderConstraintData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsConstraint
}

func TestStageLoadDebugDataBindingOrdersConstraintAfterLaterRigidBody(t *testing.T) {
	debugStageOrderFailures = nil
	key := pod.QualifiedNameForLayout(debugStageOrderConstraintData{})
	if err := engine.RegisterEntityData(debugStageOrderConstraintData{}); err != nil {
		t.Fatalf("failed to register debug constraint data: %v", err)
	}
	t.Cleanup(func() {
		delete(engine.DebugEntityDataRegistry, key)
		pod.Unregister(debugStageOrderConstraintData{})
	})
	rigidBodyBinding := EntityDataBinding{
		RegistraionKey: engine_entity_data_physics.BindingKey(),
		Fields: map[string]any{
			"Extent":   []interface{}{1.0, 1.0, 1.0},
			"IsStatic": true,
		},
	}
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "source",
				DataBinding: []EntityDataBinding{
					{
						RegistraionKey: key,
						Fields: map[string]any{
							"Target": "target",
						},
					},
					rigidBodyBinding,
				},
			},
			{
				Id: "target",
				DataBinding: []EntityDataBinding{
					rigidBodyBinding,
				},
			},
		},
	}

	host := engine.NewHost("test", nil, nil)
	stage.Load(host)

	if len(debugStageOrderFailures) > 0 {
		t.Fatalf("expected debug constraint to find staged rigid bodies, got %v", debugStageOrderFailures)
	}
	if len(host.Physics().World().Constraints()) != 1 {
		t.Fatalf("expected 1 debug-loaded constraint, got %d", len(host.Physics().World().Constraints()))
	}
}
