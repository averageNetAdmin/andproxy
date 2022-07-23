package ranges

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Create(text string) ([]string, error) {
	regrng, err := regexp.Compile(`\[\d+-\d+?\]`)

	if err != nil {
		return nil, err
	}
	rloc := regrng.FindStringIndex(text)
	if rloc == nil {
		return []string{text}, nil
	}
	res := []string{}
	ress := ""
	rng := regrng.FindString(text)

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
