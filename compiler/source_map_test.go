package compiler

import (
	"reflect"
	"testing"
)

func TestSourceMap_Add(t *testing.T) {
	type args struct {
		t         token
		destRange Range
	}
	tests := map[string]struct {
		args     args
		toSource map[int]map[int]Position
		toTarget map[int]map[int]Position
	}{
		"first line": {
			args: args{
				t: token{
					lit:  "foo",
					line: 1,
					col:  1,
				},
				destRange: Range{
					From: Position{Line: 1, Col: 1},
					To:   Position{Line: 1, Col: 3},
				},
			},
			toSource: map[int]map[int]Position{
				0: {
					0: {Line: 0, Col: 0},
					1: {Line: 0, Col: 1},
					2: {Line: 0, Col: 2},
					3: {Line: 0, Col: 3},
				},
			},
			toTarget: map[int]map[int]Position{
				0: {
					0: {Line: 0, Col: 0},
					1: {Line: 0, Col: 1},
					2: {Line: 0, Col: 2},
					3: {Line: 0, Col: 3},
				},
			},
		},
		"non-first line": {
			args: args{
				t: token{
					lit:  "foo",
					line: 2,
					col:  10,
				},
				destRange: Range{
					From: Position{Line: 5, Col: 3},
					To:   Position{Line: 5, Col: 5},
				},
			},
			toSource: map[int]map[int]Position{
				1: {
					9:  {Line: 4, Col: 2},
					10: {Line: 4, Col: 3},
					11: {Line: 4, Col: 4},
					12: {Line: 4, Col: 5},
				},
			},
			toTarget: map[int]map[int]Position{
				4: {
					2: {Line: 1, Col: 9},
					3: {Line: 1, Col: 10},
					4: {Line: 1, Col: 11},
					5: {Line: 1, Col: 12},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := &SourceMap{
				SourceLinesToTarget: make(map[int]map[int]Position),
				TargetLinesToSource: make(map[int]map[int]Position),
			}
			sm.Add(tt.args.t, tt.args.destRange)
			if !reflect.DeepEqual(tt.toSource, sm.SourceLinesToTarget) {
				t.Errorf("expected source lines to target to be %v, got %v", tt.toSource, sm.SourceLinesToTarget)
			}
			if !reflect.DeepEqual(tt.toTarget, sm.TargetLinesToSource) {
				t.Errorf("expected target lines to source to be %v, got %v", tt.toTarget, sm.TargetLinesToSource)
			}
		})
	}
}

func TestSourceMap_SourcePositionFromTarget(t *testing.T) {
	sm := &SourceMap{
		SourceLinesToTarget: map[int]map[int]Position{
			0: {
				0: {Line: 0, Col: 0},
				1: {Line: 0, Col: 1},
				2: {Line: 0, Col: 2},
				3: {Line: 0, Col: 3},
			},
			4: {
				2: {Line: 1, Col: 9},
				3: {Line: 1, Col: 10},
				4: {Line: 1, Col: 11},
				5: {Line: 1, Col: 12},
			},
		},
		TargetLinesToSource: map[int]map[int]Position{
			0: {
				0: {Line: 0, Col: 0},
				1: {Line: 0, Col: 1},
				2: {Line: 0, Col: 2},
				3: {Line: 0, Col: 3},
			},
			1: {
				9:  {Line: 4, Col: 2},
				10: {Line: 4, Col: 3},
				11: {Line: 4, Col: 4},
				12: {Line: 4, Col: 5},
			},
		},
	}
	type args struct {
		line int
		col  int
	}
	tests := map[string]struct {
		args     args
		expected Position
		ok       bool
	}{
		"first line": {
			args:     args{line: 0, col: 1},
			expected: Position{Line: 0, Col: 1},
			ok:       true,
		},
		"not found": {
			args: args{line: 1, col: 1},
		},
		"non-first line": {
			args:     args{line: 1, col: 10},
			expected: Position{Line: 4, Col: 3},
			ok:       true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := sm.SourcePositionFromTarget(tt.args.line, tt.args.col)
			if !reflect.DeepEqual(tt.expected, got) {
				t.Errorf("expected source position to be %v, got %v", tt.expected, got)
			}
			if tt.ok != ok {
				t.Errorf("expected ok to be %v, got %v", tt.ok, ok)
			}
		})
	}
}

func TestSourceMap_TargetPositionFromSource(t *testing.T) {
	sm := &SourceMap{
		SourceLinesToTarget: map[int]map[int]Position{
			0: {
				0: {Line: 0, Col: 0},
				1: {Line: 0, Col: 1},
				2: {Line: 0, Col: 2},
				3: {Line: 0, Col: 3},
			},
			4: {
				2: {Line: 1, Col: 9},
				3: {Line: 1, Col: 10},
				4: {Line: 1, Col: 11},
				5: {Line: 1, Col: 12},
			},
		},
		TargetLinesToSource: map[int]map[int]Position{
			0: {
				0: {Line: 0, Col: 0},
				1: {Line: 0, Col: 1},
				2: {Line: 0, Col: 2},
				3: {Line: 0, Col: 3},
			},
			1: {
				9:  {Line: 4, Col: 2},
				10: {Line: 4, Col: 3},
				11: {Line: 4, Col: 4},
				12: {Line: 4, Col: 5},
			},
		},
	}
	type args struct {
		line int
		col  int
	}
	tests := map[string]struct {
		args     args
		expected Position
		ok       bool
	}{
		"first line": {
			args:     args{line: 0, col: 1},
			expected: Position{Line: 0, Col: 1},
			ok:       true,
		},
		"not found": {
			args: args{line: 1, col: 1},
		},
		"non-first line": {
			args:     args{line: 4, col: 3},
			expected: Position{Line: 1, Col: 10},
			ok:       true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := sm.TargetPositionFromSource(tt.args.line, tt.args.col)
			if !reflect.DeepEqual(tt.expected, got) {
				t.Errorf("expected target position to be %v, got %v", tt.expected, got)
			}
			if tt.ok != ok {
				t.Errorf("expected ok to be %v, got %v", tt.ok, ok)
			}
		})
	}
}
