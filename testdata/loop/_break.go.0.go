package loop

import "fmt"

func breakFn() {
	k := 0

	for i := 0; i < 100; i++ {
		if i%2 == 1 {
			k += i
			break
		}
	}

	for j := 0; j < 400; j++ {
		if j%2 == 1 {
			k += j
			break
		}
	}

	fmt.Println(k)
}
