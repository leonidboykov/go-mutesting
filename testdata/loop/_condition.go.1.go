package loop

import "fmt"

func condition() {
	k := 0

	for k < 100 {
		k = k + 1
	}

	for i := 0; 1 < 1; i++ {
		k = k + 2
	}

	fmt.Println(k)
}
