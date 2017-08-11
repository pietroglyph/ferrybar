package main

import "fmt"

func (v *vesselLocation) progress() int {
	fmt.Println(v.Eta.Minute())
	return 0
}
