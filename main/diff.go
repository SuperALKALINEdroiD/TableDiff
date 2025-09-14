package main

import "fmt"

func diffStringArray(left, right []string) {
	leftSet := make(map[string]struct{})
	rightSet := make(map[string]struct{})

	for _, l := range left {
		leftSet[l] = struct{}{}
	}
	for _, r := range right {
		rightSet[r] = struct{}{}
	}

	for l := range leftSet {
		if _, ok := rightSet[l]; !ok {
			fmt.Printf("LEFT ONLY: %s\n", l)
		}
	}

	for r := range rightSet {
		if _, ok := leftSet[r]; !ok {
			fmt.Printf("RIGHT ONLY: %s\n", r)
		}
	}
}
