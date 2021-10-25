package test

import "fmt"

func print1(x int) {
	switch x {
	case 1:
		fmt.Println("1")
	case 2:
		fmt.Println("2")
	case 4:
		fmt.Println("4")
	}
}