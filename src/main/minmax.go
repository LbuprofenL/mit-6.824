package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode"
)

/*
Maximum number: 8297
Minimum number: 3
*/

func main() {
	// Read the contents of the file.
	contents, err := ioutil.ReadFile("pr.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Function to detect digit pairs.
	ff := func(r rune) bool { return !unicode.IsDigit(r) }

	// Split contents by lines.
	lines := strings.Split(string(contents), "\n")

	// Extract numbers from each line.
	var numbers []int
	for _, line := range lines {
		numStrs := strings.FieldsFunc(line, ff)
		for _, numStr := range numStrs {
			num, err := strconv.Atoi(numStr)
			if err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	// Find the maximum and minimum numbers.
	if len(numbers) == 0 {
		fmt.Println("No numbers found.")
		return
	}

	max := numbers[0]
	min := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
		if num < min {
			min = num
		}
	}

	fmt.Println("Maximum number:", max)
	fmt.Println("Minimum number:", min)
}
