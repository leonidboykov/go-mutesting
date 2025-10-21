package loop

import "fmt"

func condition() {
	k := 0

	for 1 < 1 {
		k = k + 1
	}

	for i := 0; i < 5; i++ {
		k = k + 2
	}

	fmt.Println(k)
}
