package main

// sign returns 1 if n is non-negative, -1 otherwise.
func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

// abs returns the absolute value of n.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// seq creates a sequence of numbers from min to max inclusive.
func seq(min, max int) []int {
	res := make([]int, abs(max-min)+1)
	s := sign(max - min)
	for i := range res {
		res[i] = min + i*s
	}
	return res
}
