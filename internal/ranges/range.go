package ranges

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// find all ranges and replase they to several lines
// return one string
//
func Create(text string) ([]string, error) {
	// regular expression to range 
	// EXAMPLE: [1-5]
	regrng, err := regexp.Compile(`\[\d+-\d+?\]`)
	if err != nil {
		return nil, err
	}
	// find range index
	rloc := regrng.FindStringIndex(text)
	if rloc == nil {
		return []string{text}, nil
	}
	res := []string{}
	ress := ""
	//find ranges
	rng := regrng.FindString(text)

	// create array from range
	rng = rng[1 : len(rng)-1]
	start, err := strconv.Atoi(strings.Split(rng, "-")[0])
	if err != nil {
		return nil, err
	}
	end, err := strconv.Atoi(strings.Split(rng, "-")[1])
	if err != nil {
		return nil, err
	}
	if start > end {
		return nil, fmt.Errorf("error parsing range: start number bigger then end")
	}
	// create many expressions from range array
	for i := start; i <= end; i++ {
		ress = text[:rloc[0]] + strconv.Itoa(i) + text[rloc[1]:]

		rs, err := Create(ress)
		if err != nil {
			return nil, err
		}
		res = append(res, rs...)

	}
	return res, nil
}
