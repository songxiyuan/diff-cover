package song_print

import (
	"fmt"
)

func Print(x int) bool {
	switch x {
	case 1:
		fmt.Println(1)
	case 2:
		fmt.Println(2)
	case 3:
		fmt.Println(3)
	}
	return false
}