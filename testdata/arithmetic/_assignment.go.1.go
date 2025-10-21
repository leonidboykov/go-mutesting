package arithmetic

import "fmt"

func assignment() {
	i := 100

	i += 10
	i = 20
	i *= 2
	i /= 2
	i %= 10000

	i &= 1
	i |= 1
	i ^= 1
	i <<= 1
	i >>= 1
	i &^= 1

	fmt.Println(i)
}
