package main

import "math"

func add(a, b interface{}) interface{} {
	if isFloat(a) || isFloat(b) {
		return toFloat(a) + toFloat(b)
	}
	return a.(int) + b.(int)
}

func subtract(a, b interface{}) interface{} {
	if isFloat(a) || isFloat(b) {
		return toFloat(a) - toFloat(b)
	}
	return a.(int) - b.(int)
}

func multiply(a, b interface{}) interface{} {
	if isFloat(a) || isFloat(b) {
		return toFloat(a) * toFloat(b)
	}
	return a.(int) * b.(int)
}

func divide(a, b interface{}) interface{} {
	if isFloat(a) || isFloat(b) {
		return toFloat(a) / toFloat(b)
	}
	return a.(int) / b.(int)
}

func modulo(a, b int) int {
	return a % b
}

func power(a, b interface{}) float64 {
	return math.Pow(toFloat(a), toFloat(b))
}

func equals(a, b interface{}) int {
	return btoi(a == b)
}

func notEquals(a, b interface{}) int {
	return btoi(a != b)
}

func lessThan(a, b interface{}) int {
	return btoi(toFloat(a) < toFloat(b))
}

func lessThanOrEqual(a, b interface{}) int {
	return btoi(toFloat(a) <= toFloat(b))
}

func greaterThan(a, b interface{}) int {
	return btoi(toFloat(a) > toFloat(b))
}

func greaterThanOrEqual(a, b interface{}) int {
	return btoi(toFloat(a) >= toFloat(b))
}

func and(a, b interface{}) int {
	return btoi(toFloat(a) != 0 && toFloat(b) != 0)
}

func or(a, b interface{}) int {
	return btoi(toFloat(a) != 0 || toFloat(b) != 0)
}

func isFloat(x interface{}) bool {
	_, ok := x.(float64)
	return ok
}

func toFloat(x interface{}) float64 {
	if f, ok := x.(float64); ok {
		return f
	}
	return float64(x.(int))
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}


func bitwiseAnd(a, b int) int {
    return a & b
}

func bitwiseOr(a, b int) int {
    return a | b
}

func bitwiseXor(a, b int) int {
    return a ^ b
}

func leftShift(a, b int) int {
    return a << uint(b)
}

func rightShift(a, b int) int {
    return a >> uint(b)
}

