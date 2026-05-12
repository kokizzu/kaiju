/******************************************************************************/
/* sample_stages.go                                                           */
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

package samples

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
)

const (
	SampleStageDistanceChain = "physics_distance_chain"
	SampleStageRope          = "physics_rope"
	SampleStageBridge        = "physics_bridge"
	SampleStageHingePendulum = "physics_hinge_pendulum"
	SampleStageBodyWorld     = "physics_body_world_anchor"
)

func ConstraintSampleStages() map[string]stages.Stage {
	return map[string]stages.Stage{
		SampleStageDistanceChain: DistanceChainSampleStage(),
		SampleStageRope:          RopeSampleStage(),
		SampleStageBridge:        BridgeSampleStage(),
		SampleStageHingePendulum: HingePendulumSampleStage(),
		SampleStageBodyWorld:     BodyWorldAnchorSampleStage(),
	}
}

func DistanceChainSampleStage() stages.Stage {
	positions := []matrix.Vec3{
		matrix.NewVec3(-3, 0, 0),
		matrix.NewVec3(-2, 0, 0),
		matrix.NewVec3(-1, 0, 0),
		matrix.NewVec3(0, 0, 0),
		matrix.NewVec3(1, 0, 0),
		matrix.NewVec3(2, 0, 0),
		matrix.NewVec3(3, 0, 0),
	}
	stage := sampleBodyChainStage(SampleStageDistanceChain, "chain", positions, func(i int) engine_entity_data_physics.RigidBodyEntityData {
		body := sampleSphereBody(false)
		body.IsStatic = i == 0 || i == len(positions)-1
		return body
	})
	for i := 0; i < len(stage.Entities)-1; i++ {
		addSampleBinding(&stage.Entities[i], engine_entity_data_physics.DistanceJointEntityData{
			ConnectedEntityId: engine.EntityId(stage.Entities[i+1].Id),
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.8,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			AutoRestLength:    true,
		})
	}
	return stage
}

func RopeSampleStage() stages.Stage {
	positions := []matrix.Vec3{
		matrix.NewVec3(0, 0, 0),
		matrix.NewVec3(0, -1, 0),
		matrix.NewVec3(0, -2, 0),
		matrix.NewVec3(0, -3, 0),
		matrix.NewVec3(0, -4, 0),
	}
	stage := sampleBodyChainStage(SampleStageRope, "rope", positions, func(i int) engine_entity_data_physics.RigidBodyEntityData {
		body := sampleSphereBody(false)
		body.IsStatic = i == 0
		return body
	})
	for i := 0; i < len(stage.Entities)-1; i++ {
		addSampleBinding(&stage.Entities[i], engine_entity_data_physics.RopeJointEntityData{
			ConnectedEntityId: engine.EntityId(stage.Entities[i+1].Id),
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.8,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			AutoMaxLength:     false,
			MaxLength:         1,
		})
	}
	return stage
}

func BridgeSampleStage() stages.Stage {
	positions := []matrix.Vec3{
		matrix.NewVec3(-3, 0, 0),
		matrix.NewVec3(-2, 0, 0),
		matrix.NewVec3(-1, 0, 0),
		matrix.NewVec3(0, 0, 0),
		matrix.NewVec3(1, 0, 0),
		matrix.NewVec3(2, 0, 0),
		matrix.NewVec3(3, 0, 0),
	}
	stage := sampleBodyChainStage(SampleStageBridge, "bridge", positions, func(i int) engine_entity_data_physics.RigidBodyEntityData {
		body := sampleBoxBody(false)
		body.IsStatic = i == 0 || i == len(positions)-1
		return body
	})
	for i := 0; i < len(stage.Entities)-1; i++ {
		addSampleBinding(&stage.Entities[i], engine_entity_data_physics.DistanceJointEntityData{
			ConnectedEntityId: engine.EntityId(stage.Entities[i+1].Id),
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.9,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			AutoRestLength:    false,
			RestLength:        1,
		})
	}
	return stage
}

func HingePendulumSampleStage() stages.Stage {
	anchor := sampleEntity("hinge_anchor", "Hinge Anchor", matrix.Vec3Zero(), sampleSphereBody(true))
	arm := sampleEntity("hinge_arm", "Hinge Arm", matrix.NewVec3(0, -2, 0), sampleBoxBody(false))
	addSampleBinding(&anchor, engine_entity_data_physics.HingeJointEntityData{
		ConnectedEntityId: engine.EntityId(arm.Id),
		LocalAnchorA:      matrix.Vec3Zero(),
		TargetAnchorB:     matrix.NewVec3(0, 2, 0),
		Stiffness:         1,
		Bias:              0.2,
		Correction:        0.8,
		Slop:              0.001,
		MaxCorrection:     0.5,
		Enabled:           true,
		HingeAxis:         matrix.Vec3Backward(),
	})
	return stages.Stage{
		Id:       SampleStageHingePendulum,
		Entities: []stages.EntityDescription{anchor, arm},
	}
}

func BodyWorldAnchorSampleStage() stages.Stage {
	body := sampleEntity("anchored_body", "Body World Anchor", matrix.NewVec3(0, -1.5, 0), sampleSphereBody(false))
	addSampleBinding(&body, engine_entity_data_physics.DistanceJointEntityData{
		LocalAnchorA:   matrix.Vec3Zero(),
		TargetAnchorB:  matrix.Vec3Zero(),
		Stiffness:      1,
		Bias:           0.2,
		Correction:     0.8,
		Slop:           0.001,
		MaxCorrection:  0.5,
		Enabled:        true,
		AutoRestLength: false,
		RestLength:     1.5,
	})
	return stages.Stage{
		Id:       SampleStageBodyWorld,
		Entities: []stages.EntityDescription{body},
	}
}

func sampleBodyChainStage(id, prefix string, positions []matrix.Vec3, bodyAt func(int) engine_entity_data_physics.RigidBodyEntityData) stages.Stage {
	stage := stages.Stage{
		Id:       id,
		Entities: make([]stages.EntityDescription, len(positions)),
	}
	for i, pos := range positions {
		entityId := fmt.Sprintf("%s_%02d", prefix, i)
		stage.Entities[i] = sampleEntity(entityId, entityId, pos, bodyAt(i))
	}
	return stage
}

func sampleEntity(id, name string, position matrix.Vec3, body engine_entity_data_physics.RigidBodyEntityData) stages.EntityDescription {
	desc := stages.EntityDescription{
		Id:       id,
		Name:     name,
		Position: position,
		Scale:    matrix.Vec3One(),
	}
	addSampleBinding(&desc, body)
	return desc
}

func sampleSphereBody(static bool) engine_entity_data_physics.RigidBodyEntityData {
	return engine_entity_data_physics.RigidBodyEntityData{
		Mass:     1,
		Radius:   0.25,
		Shape:    engine_entity_data_physics.ShapeSphere,
		IsStatic: static,
	}
}

func sampleBoxBody(static bool) engine_entity_data_physics.RigidBodyEntityData {
	return engine_entity_data_physics.RigidBodyEntityData{
		Extent:   matrix.NewVec3(0.45, 0.12, 0.12),
		Mass:     1,
		Radius:   0.25,
		Height:   0.25,
		Shape:    engine_entity_data_physics.ShapeBox,
		IsStatic: static,
	}
}

func addSampleBinding[T any](desc *stages.EntityDescription, data T) {
	desc.RawDataBinding = append(desc.RawDataBinding, data)
	desc.DataBinding = append(desc.DataBinding, stages.EntityDataBinding{
		RegistraionKey: pod.QualifiedNameForLayout(data),
		Fields:         sampleBindingFields(data),
	})
}

func sampleBindingFields(data any) map[string]any {
	switch v := data.(type) {
	case engine_entity_data_physics.RigidBodyEntityData:
		return map[string]any{
			"AssetKey": v.AssetKey,
			"Extent":   v.Extent,
			"Mass":     v.Mass,
			"Radius":   v.Radius,
			"Height":   v.Height,
			"Shape":    v.Shape,
			"IsStatic": v.IsStatic,
		}
	case engine_entity_data_physics.DistanceJointEntityData:
		return map[string]any{
			"ConnectedEntityId": v.ConnectedEntityId,
			"LocalAnchorA":      v.LocalAnchorA,
			"TargetAnchorB":     v.TargetAnchorB,
			"Stiffness":         v.Stiffness,
			"Bias":              v.Bias,
			"Correction":        v.Correction,
			"Slop":              v.Slop,
			"MaxCorrection":     v.MaxCorrection,
			"WarmStarting":      v.WarmStarting,
			"Enabled":           v.Enabled,
			"BreakForce":        v.BreakForce,
			"BreakTorque":       v.BreakTorque,
			"RestLength":        v.RestLength,
			"AutoRestLength":    v.AutoRestLength,
		}
	case engine_entity_data_physics.RopeJointEntityData:
		return map[string]any{
			"ConnectedEntityId": v.ConnectedEntityId,
			"LocalAnchorA":      v.LocalAnchorA,
			"TargetAnchorB":     v.TargetAnchorB,
			"Stiffness":         v.Stiffness,
			"Bias":              v.Bias,
			"Correction":        v.Correction,
			"Slop":              v.Slop,
			"MaxCorrection":     v.MaxCorrection,
			"WarmStarting":      v.WarmStarting,
			"Enabled":           v.Enabled,
			"BreakForce":        v.BreakForce,
			"BreakTorque":       v.BreakTorque,
			"MaxLength":         v.MaxLength,
			"AutoMaxLength":     v.AutoMaxLength,
		}
	case engine_entity_data_physics.HingeJointEntityData:
		return map[string]any{
			"ConnectedEntityId": v.ConnectedEntityId,
			"LocalAnchorA":      v.LocalAnchorA,
			"TargetAnchorB":     v.TargetAnchorB,
			"Stiffness":         v.Stiffness,
			"Bias":              v.Bias,
			"Correction":        v.Correction,
			"Slop":              v.Slop,
			"MaxCorrection":     v.MaxCorrection,
			"WarmStarting":      v.WarmStarting,
			"Enabled":           v.Enabled,
			"BreakForce":        v.BreakForce,
			"BreakTorque":       v.BreakTorque,
			"HingeAxis":         v.HingeAxis,
			"EnableLimits":      v.EnableLimits,
			"MinAngleDegrees":   v.MinAngleDegrees,
			"MaxAngleDegrees":   v.MaxAngleDegrees,
			"EnableMotor":       v.EnableMotor,
			"MotorSpeedDegrees": v.MotorSpeedDegrees,
			"MaxMotorTorque":    v.MaxMotorTorque,
			"MaxMotorImpulse":   v.MaxMotorImpulse,
		}
	default:
		return map[string]any{}
	}
}
