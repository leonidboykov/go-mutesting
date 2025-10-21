package arithmetic

import "fmt"

func assign() {
	i := 100

	i += 10
	i -= 20
	i *= 2
	i *= 5
	i %= 10

	fmt.Println(i)
}
