package test

import (
	"fmt"
	"regexp"
	"testing"
)

func TestBlindBoxRegex(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"3月盲盒", true},
		{"2024年3月盲盒", true},
		{"3月心动盲盒", true},
		{"2024年3月心动盲盒", true},
	}

	reg1 := `^(?:([0-9]{4})年)?([0-9]+)月盲盒$`
	reg2 := `^(?:([0-9]{4})年)?([0-9]+)月([^盲]+)盲盒$`

	for _, tc := range testCases {
		re1 := regexp.MustCompile(reg1)
		re2 := regexp.MustCompile(reg2)
		
		match1 := re1.FindStringSubmatch(tc.input)
		match2 := re2.FindStringSubmatch(tc.input)
		
		fmt.Printf("测试输入: %s\n", tc.input)
		fmt.Printf("普通盲盒匹配结果: %v\n", match1)
		fmt.Printf("指定类型盲盒匹配结果: %v\n", match2)
		fmt.Println("-------------------")
	}
}