/******************************************************************************/
/* physics_system.go                                                          */
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

package engine

import (
	"log/slog"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
)

type StagePhysicsEntry struct {
	Entity *Entity
	Body   *graviton.RigidBody
}

type StagePhysics struct {
	world    graviton.System
	entities []StagePhysicsEntry
	active   bool
}

func (pe *StagePhysicsEntry) syncEntityToBody() {
	t := &pe.Entity.Transform
	b := pe.Body
	b.Transform.SetPosition(t.WorldPosition())
	b.Transform.SetRotation(t.WorldRotation())
	b.Transform.SetScale(t.WorldScale())
}

func (pe *StagePhysicsEntry) syncBodyToEntity() {
	b := pe.Body
	t := &pe.Entity.Transform
	t.SetWorldPosition(b.Position())
	t.SetWorldRotation(b.Rotation().ToEuler())
	t.SetWorldScale(b.Transform.WorldScale())
}

func (p *StagePhysics) IsActive() bool          { return p.active }
func (p *StagePhysics) World() *graviton.System { return &p.world }

func (p *StagePhysics) Start() {
	defer tracing.NewRegion("StagePhysics.StagePhysics").End()
	if p.active {
		slog.Error("Stage physics has already started, can not start again")
		return
	}
	p.world.Initialize()
	p.world.SetGravity(matrix.NewVec3(0, -9.81, 0))
	p.active = true
}

func (p *StagePhysics) Destroy() {
	defer tracing.NewRegion("StagePhysics.Destroy").End()
	if p.active {
		p.world.Clear()
	}
	p.entities = klib.WipeSlice(p.entities)
	p.active = false
}

func (p *StagePhysics) AddEntity(entity *Entity, body *graviton.RigidBody) {
	defer tracing.NewRegion("StagePhysics.AddEntity").End()
	if !p.active {
		slog.Error("stage physics has not started, can not add entity")
		return
	}
	if entity == nil || body == nil {
		slog.Error("failed to add entity physics, entity and body are required")
		return
	}
	body.Transform.SetPosition(entity.Transform.WorldPosition())
	body.Transform.SetRotation(entity.Transform.WorldRotation())
	body.Transform.SetScale(entity.Transform.WorldScale())
	stageBody := p.world.AddBody(body)
	if stageBody == nil {
		slog.Error("failed to add entity physics body")
		return
	}
	p.entities = append(p.entities, StagePhysicsEntry{
		Entity: entity,
		Body:   stageBody,
	})
	entity.OnDestroy.Add(func() {
		cIdx := -1
		for i := range p.entities {
			if p.entities[i].Entity == entity {
				cIdx = i
				break
			}
		}
		if cIdx != -1 {
			p.entities = klib.RemoveUnordered(p.entities, cIdx)
			p.world.RemoveBody(stageBody)
		}
	})
}

func (p *StagePhysics) AddEntityShape(entity *Entity, mass float32, shape graviton.Shape) {
	defer tracing.NewRegion("StagePhysics.AddEntityShape").End()
	t := &entity.Transform
	inertia := graviton.CalculateLocalInertia(shape, matrix.Float(mass))
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(t.Position())
	body.Transform.SetRotation(t.Rotation())
	body.SetShape(shape)
	if mass <= 0 {
		body.SetStatic()
	} else {
		body.SetDynamic(matrix.Float(mass), inertia)
	}
	p.AddEntity(entity, body)
}

func (p *StagePhysics) Update(workGroup *concurrent.WorkGroup, threads *concurrent.Threads, deltaTime float64) {
	defer tracing.NewRegion("StagePhysics.Update").End()
	for i := range p.entities {
		entry := &p.entities[i]
		if entry.Body.IsKinematic() || (entry.Body.IsStatic() && entry.Entity.Transform.IsDirty()) {
			entry.syncEntityToBody()
		}
	}
	p.world.Step(workGroup, threads, deltaTime)
	for i := range p.entities {
		if p.entities[i].Body.IsDynamic() {
			p.entities[i].syncBodyToEntity()
		}
	}
}
