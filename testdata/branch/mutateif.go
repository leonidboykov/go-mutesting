package branch

import (
	"fmt"
)

func mutateif() {
	i := 1

	for i != 4 {
		if i == 1 {
			fmt.Println(i)
		} else if i == 2 {
			fmt.Println(i * 2)
		} else {
			fmt.Println(i * 3)
		}

		i++
	}
}
