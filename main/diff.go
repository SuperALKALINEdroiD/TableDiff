package main

import "fmt"

func diffColumns(left, right []ColumnInfo) {
	leftMap := make(map[string]ColumnInfo, len(left))
	rightMap := make(map[string]ColumnInfo, len(right))

	for _, col := range left {
		leftMap[col.Name] = col
	}
	for _, col := range right {
		rightMap[col.Name] = col
	}

	for name, leftCol := range leftMap {
		if rightCol, ok := rightMap[name]; ok {
			if leftCol.Type != rightCol.Type {
				fmt.Printf("TYPE DIFF: %s - LEFT: %s, RIGHT: %s\n", name, leftCol.Type, rightCol.Type)
			}
			if leftCol.Default != rightCol.Default {
				leftDefault := leftCol.Default
				rightDefault := rightCol.Default
				if leftDefault == "" {
					leftDefault = "NULL"
				}
				if rightDefault == "" {
					rightDefault = "NULL"
				}
				fmt.Printf("DEFAULT DIFF: %s - LEFT: %s, RIGHT: %s\n", name, leftDefault, rightDefault)
			}
		} else {
			defaultStr := leftCol.Default
			if defaultStr == "" {
				defaultStr = "NULL"
			}
			fmt.Printf("LEFT ONLY: %s (%s, default: %s)\n", name, leftCol.Type, defaultStr)
		}
	}

	for name, col := range rightMap {
		if _, ok := leftMap[name]; !ok {
			defaultStr := col.Default
			if defaultStr == "" {
				defaultStr = "NULL"
			}
			fmt.Printf("RIGHT ONLY: %s (%s, default: %s)\n", name, col.Type, defaultStr)
		}
	}
}
