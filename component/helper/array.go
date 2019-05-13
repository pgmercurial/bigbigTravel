package helper

import (
	"math/rand"
	"time"
)

func ArrayIntIntersect(slice ...[]int) []int {
	if len(slice) <= 1 {
		return slice[0]
	}

	result := make([]int, 0, len(slice[0]))
	for _, v := range slice[0] {
		allExists := true
		for i := 1; i < len(slice); i++ {
			exists := false
			for _, v2 := range slice[i] {
				if v == v2 {
					exists = true
					break
				}
			}
			if !exists {
				allExists = false
				break
			}
		}
		if allExists {
			result = append(result, v)
		}
	}
	return result
}

func ArrayStringIntersect(slice ...[]string) []string {
	if len(slice) <= 1 {
		return slice[0]
	}

	result := make([]string, 0, len(slice[0]))
	for _, v := range slice[0] {
		allExists := true
		for i := 1; i < len(slice); i++ {
			exists := false
			for _, v2 := range slice[i] {
				if v == v2 {
					exists = true
					break
				}
			}
			if !exists {
				allExists = false
				break
			}
		}
		if allExists {
			result = append(result, v)
		}
	}
	return result
}

func UniqueInts(input []int) []int {
	u := make([]int, 0, len(input))
	m := make(map[int]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

func UniqueStrings(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

func SliceSearchInt(slice []int, int int) int {
	for i, v := range slice {
		if v == int {
			return i
		}
	}
	return -1
}

func SliceIntShuffle(slice []int) []int {
	l := len(slice)
	rr := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	for i := l - 1; i > 0; i-- {
		r := rr.Intn(i)
		slice[r], slice[i] = slice[i], slice[r]
	}
	return slice
}

func RemoveListRepeatedPart(fullSlice, removeSlice []int) []int {
	removeMap := make(map[int]bool)
	for _, removeElement := range removeSlice {
		removeMap[removeElement] = true
	}

	result := make([]int, 0)
	for _, fullElement := range fullSlice {
		if _, ok := removeMap[fullElement]; !ok {
			result = append(result, fullElement)
		}
	}

	return result
}

func StringInSlice(slice []string, key string) bool {
	for _, v := range slice {
		if key == v {
			return true
		}
	}

	return false
}

func IntInSlice(slice []int, key int) bool {
	for _, v := range slice {
		if key == v {
			return true
		}
	}

	return false
}
