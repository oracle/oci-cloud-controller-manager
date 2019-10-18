// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import "sort"

// permissionsIntersect returns the intersect (A ∩ B) of the two given permission lists
func permissionsIntersect(pa []string, pb []string) []string {
	// we use a map to fake a set
	result := make(map[string]bool)

	// A ∩ B
	for _, pai := range pa {
		for _, pbi := range pb {
			if pai == pbi {
				result[pai] = true
				continue
			}
		}
	}

	// map -> array
	set := make([]string, 0, len(result))
	for k := range result {
		set = append(set, k)
	}

	// Sort so that we have consistent return value
	sort.Strings(set)

	return set
}

// permissionDifference returns the difference (A - B) of two sets of permission strings; e.g. all the items in A
// not in B
func permissionsDifference(pa []string, pb []string) []string {
	// We use a map to make a fake set
	result := make(map[string]bool)

	// A - B
	for _, pai := range pa {
		found := false
		for _, pbi := range pb {
			if pai == pbi {
				found = true
				break
			}
		}
		if !found {
			result[pai] = true
		}
	}

	// map -> array
	set := make([]string, 0, len(result))
	for k := range result {
		set = append(set, k)
	}

	// Sort so that we have consistent return value
	sort.Strings(set)

	return set
}

// getPermissionsFromContextVariable returns a list of permissions from the given ContextVariables
func getPermissionsFromContextVariables(c [][]ContextVariable) []string {
	result := make([]string, 0, len(c))
	for _, ci := range c {
		if ci[0].P != "" && ci[0].P != commonPermission {
			result = append(result, ci[0].P)
		}
	}
	return result
}
