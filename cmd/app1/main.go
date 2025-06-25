package main

import "fmt"

type Person struct {
	Name string
}

func main() {
	// slices
	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	fmt.Println(double(list))

	// work with pointers
	person := &Person{}
	person.Name = "Bob"
	fmt.Println(person.Name)
	changeName(person)
	fmt.Println(person.Name)

	// map, карта
	profile := map[int]string{1: "john", 2: "ivan"}

	// цикл итерации
	for k, v := range profile {
		fmt.Println(k, v)
	}
}

// return value * 2
func double(nums []int) []int {
	result := make([]int, 0, len(nums))

	for _, n := range nums {
		result = append(result, n*2)
	}
	return result
}

func changeName(person *Person) {
	// person = &Person{}
	person.Name = "Alice"
}
