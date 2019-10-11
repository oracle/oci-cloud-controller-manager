// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPermissionsIntersect(t *testing.T) {
	testIO := []struct {
		tc       string
		a        []string
		b        []string
		expected []string
	}{
		{
			tc:       `should return empty list when both inputs are empty`,
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			tc:       `should return empty list when the first input is empty`,
			a:        []string{"a", "b", "c"},
			b:        []string{},
			expected: []string{},
		},
		{
			tc:       `should return empty list when the second input is empty`,
			a:        []string{},
			b:        []string{"b", "c", "d"},
			expected: []string{},
		},
		{
			tc:       `should return expected set when both inputs are equal`,
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			tc:       `should return expected intersect of the two inputs`,
			a:        []string{"a", "c", "d", "f"},
			b:        []string{"a", "b", "d", "e"},
			expected: []string{"a", "d"},
		},
		{
			tc:       `should return expected subset set of first input`,
			a:        []string{"a", "b", "c"},
			b:        []string{"a"},
			expected: []string{"a"},
		},
		{
			tc:       `should return expected set of second input`,
			a:        []string{"a"},
			b:        []string{"a", "b", "c"},
			expected: []string{"a"},
		},
		{
			tc:       `should return expected set of first input with duplicate entries`,
			a:        []string{"a", "a", "a"},
			b:        []string{"a", "b"},
			expected: []string{"a"},
		},
		{
			tc:       `should return expected set of second input with duplicate entries`,
			a:        []string{"a", "b"},
			b:        []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			tc:       `should return expected set of when both sets contain duplicate entries`,
			a:        []string{"a", "a", "a", "d", "c", "d", "c"},
			b:        []string{"a", "b", "a", "b", "b", "c"},
			expected: []string{"a", "c"},
		},
		{
			tc:       `should return an empty set when there is no intersection`,
			a:        []string{"a", "b", "c", "d", "e"},
			b:        []string{"f", "g", "h", "i", "j"},
			expected: []string{},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			r := permissionsIntersect(test.a, test.b)
			assert.Equal(t, r, test.expected)
		})
	}
}

func TestGetPermissionsDifference(t *testing.T) {
	testIO := []struct {
		tc       string
		a        []string
		b        []string
		expected []string
	}{
		{
			tc:       `should return empty list when both inputs are empty`,
			a:        []string{},
			b:        []string{},
			expected: []string{},
		},
		{
			tc:       `should return all items in first input if the second is empty`,
			a:        []string{"a", "b", "c"},
			b:        []string{},
			expected: []string{"a", "b", "c"},
		},
		{
			tc:       `should return empty set if the first set is empty`,
			a:        []string{},
			b:        []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			tc:       `should return empty set when the two sets are equal`,
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: []string{},
		},
		{
			tc:       `should return the difference when the two sets are not equal`,
			a:        []string{"a", "c", "d"},
			b:        []string{"a", "b", "d"},
			expected: []string{"c"},
		},
		{
			tc:       `should return the difference in first set`,
			a:        []string{"a", "b", "c"},
			b:        []string{"a"},
			expected: []string{"b", "c"},
		},
		{
			tc:       `should empty set if the first set is a subset of the second`,
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c", "d", "e"},
			expected: []string{},
		},
		{
			tc:       `should return expected set when the first set contains duplicates`,
			a:        []string{"a", "a", "c", "c", "a"},
			b:        []string{"a", "b"},
			expected: []string{"c"},
		},
	}

	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			r := permissionsDifference(test.a, test.b)
			assert.Equal(t, r, test.expected)
		})
	}
}

func TestPermissionFromContextVariable(t *testing.T) {
	testIO := []struct {
		tc       string
		input    [][]ContextVariable
		expected []string
	}{
		{
			tc:       `should return empty permission list from empty context variables`,
			input:    [][]ContextVariable{},
			expected: []string{},
		},
		{
			tc:       `should return expected permission list from context variables`,
			input:    [][]ContextVariable{{{P: `permission-1`}}},
			expected: []string{`permission-1`},
		},
		{
			tc: `should return expected permission list from multiple context variables`,
			input: [][]ContextVariable{
				{{P: `permission-1`}},
				{{P: `permission-2`}},
			},
			expected: []string{`permission-1`, `permission-2`},
		},
		{
			tc: `should only return the first permission if the context record contains more than one permission`,
			input: [][]ContextVariable{
				{{P: `permission-1`}, {P: `permission-2`}},
			},
			expected: []string{`permission-1`},
		},
	}
	for _, test := range testIO {
		t.Run(test.tc, func(t *testing.T) {
			r := getPermissionsFromContextVariables(test.input)
			assert.Equal(t, r, test.expected)
		})
	}
}
