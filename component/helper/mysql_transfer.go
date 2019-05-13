package helper

import (
	"errors"
	"strconv"
)

func IntList2MysqlWhereIn(intList []int) (string, error) {
	var result = ""
	len := len(intList)
	if len == 0 {
		return result, errors.New("length of intList cannot be 0")
	}
	result = result + "("
	for index, element := range intList {
		result = result + strconv.Itoa(element)
		if index < len-1 {
			result = result + ","
		}
	}
	result = result + ")"
	return result, nil
}
