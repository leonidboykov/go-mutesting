package arithmetic

import "fmt"

func base() {
	i := 100

	i = i + 10
	i = i - 20
	i = i * 2
	i = i / 2
	i = i % 10

	fmt.Println(i)
}
