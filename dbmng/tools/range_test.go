package tools

import (
	"testing"
)

// find all ranges and replase they to several lines
// return one string
func TestCreateRange(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		res, _ := CreateRange("172.16.[1-4].0/23")
		want := []string{
			"172.16.1.0/23",
			"172.16.2.0/23",
			"172.16.3.0/23",
			"172.16.4.0/23",
		}
		if !compareStringArray(res, want) {
			t.Errorf("WANT %s\n GOT %s\n", want, res)
		}
	})
	t.Run("Success", func(t *testing.T) {
		res, _ := CreateRange("172.[1-2].[1-3].0/23")
		want := []string{
			"172.1.1.0/23",
			"172.1.2.0/23",
			"172.1.3.0/23",
			"172.2.1.0/23",
			"172.2.2.0/23",
			"172.2.3.0/23",
		}
		if !compareStringArray(res, want) {
			t.Errorf("WANT %s\n GOT %s\n", want, res)
		}
	})
	// t.Run("Error", func(t *testing.T) {
	// 	res, _ := CreateRange("172.16.[-4].0/23")
	// 	want := []string{
	// 		"172.16.1.0/23",
	// 		"172.16.2.0/23",
	// 		"172.16.3.0/23",
	// 		"172.16.4.0/23",
	// 	}
	// 	if !compareStringArray(res, want) {
	// 		t.Errorf("WANT %s\n GOT %s\n", want, res)
	// 	}
	// })
}

func compareStringArray(a1, a2 []string) bool {
	var equal bool = false
	for _, s1 := range a1 {
		for _, s2 := range a2 {
			if s1 == s2 {
				equal = true
				break
			}
		}
		if !equal {
			return false
		}
		equal = false
	}
	return true
}
