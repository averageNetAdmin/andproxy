package ippool

import (
	"regexp"
	"strconv"
	"strings"
)

//
//	create range. for example srv1-3 will converted to [srv1, srv2, srv3]
//
func CreateRange(rng string) ([]string, error) {

	reg, err := regexp.Compile(`\d+?-\d+`)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	rn := reg.FindString(rng)

	loc := reg.FindStringIndex(rng)

	nums := strings.Split(rn, "-")
	start, _ := strconv.Atoi(nums[0])
	end, _ := strconv.Atoi(nums[1])
	for i := start; i <= end; i++ {
		l := rng[:loc[0]] + strconv.Itoa(i) + rng[loc[1]:]
		res = append(res, l)
	}

	return res, nil
}
