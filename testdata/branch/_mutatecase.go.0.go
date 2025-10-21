package branch

import (
	"fmt"
)

func mutatecase() {
	i := 1

	for i != 4 {
		switch {
		case i == 1:
			_, _ = fmt.Println, i
		case i == 2:
			fmt.Println(i * 2)
		default:
			fmt.Println(i * 3)
		}

		i++
	}
}
