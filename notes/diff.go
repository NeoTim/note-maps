// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notes

import (
	"github.com/google/note-maps/notes/change"
)

// Stage describes a set of changes that might be made to a note map.
//
// The default stage describes an empty set of changes to be made to an empty
// note map.
//
// A default Stage{} is an empty set of changes made to an empty note map.
type Stage struct {
	Ops  []change.Operation
	Base Loader
}

// Add simply appends o to the set of operations described by x.
func (x *Stage) Add(o change.Operation) *Stage {
	x.Ops = append(x.Ops, o)
	return x
}

// Note returns a note-specific StageNote focused on note with id.
func (x *Stage) Note(id uint64) *StageNote { return &StageNote{x, id} }

// GetBase returns a non-nil Loader derived from x.Base.
func (x *Stage) GetBase(defaultEmpty bool) Loader {
	base := x.Base
	if base == nil {
		base = EmptyLoader
	}
	return base
}

// StageNote supports updating the content of a note within a batch, and also
// implements the Note interface to read the hypothetical state of a note with
// the batch applied.
type StageNote struct {
	Stage *Stage
	Id    uint64
}

func (x *StageNote) GetId() uint64 { return x.Id }
func (x *StageNote) GetTypes() ([]Note, error) {
	base, err := LoadOne(x.Stage.GetBase(true), x.Id)
	if err != nil {
		return nil, err
	}
	return base.GetTypes()
}
func (x *StageNote) GetSupertypes() ([]Note, error) {
	base, err := LoadOne(x.Stage.GetBase(true), x.Id)
	if err != nil {
		return nil, err
	}
	return base.GetSupertypes()
}
func (x *StageNote) GetValue() (string, Note, error) {
	base, err := LoadOne(x.Stage.GetBase(true), x.Id)
	if err != nil {
		return "", nil, err
	}
	lex, dtype, err := base.GetValue()
	if err != nil {
		return lex, dtype, err
	}
	for _, op := range x.Stage.Ops {
		if op.AffectsId(x.Id) {
			switch o := op.(type) {
			case change.SetValue:
				lex, dtype = o.Lexical, x.Stage.Note(o.Datatype)
			}
		}
	}
	return lex, dtype, nil
}
func (x *StageNote) GetContents() ([]Note, error) {
	base, err := LoadOne(x.Stage.GetBase(true), x.Id)
	if err != nil {
		return nil, err
	}
	ns, err := base.GetContents()
	if err != nil {
		return ns, err
	}
	for _, op := range x.Stage.Ops {
		if op.AffectsId(x.Id) {
			switch o := op.(type) {
			case change.AddContent:
				ns = append(ns, x.Stage.Note(o.Add))
			}
		}
	}
	return ns, nil
}

// SetValue expands the staged operations to update the value of this note.
func (x *StageNote) SetValue(lexical string, datatype uint64) {
	if x.Id == 0 {
		panic("id is still 0")
	}
	x.Stage.Add(change.SetValue{Id: x.Id, Lexical: lexical, Datatype: datatype})
}

// AddContent expands the staged operations to add content to this note.
func (x *StageNote) AddContent(id uint64) *StageNote {
	if x.Id == 0 {
		panic("id is still 0")
	}
	x.Stage.Add(change.AddContent{Id: x.Id, Add: id})
	return &StageNote{x.Stage, id}
}
