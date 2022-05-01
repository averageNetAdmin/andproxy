package ippool

import (
	"net/netip"
	"strconv"
	"strings"
)

//
// create array of IP addresses from IP range
// for example "192.168.0-1.1-5" will converted to
// [192.168.0.1 192.168.0.2 192.168.0.3 192.168.0.4 192.168.0.5
// 192.168.1.1 192.168.1.2 192.168.1.3 192.168.1.4 192.168.1.5]
//
func CreateIPRange(iprange string) ([]netip.Addr, error) {

	octet := 0
	result := make([]netip.Addr, 0)
	splittedIP := strings.Split(iprange, ".")

	var f func([]string) error
	f = func(sipr []string) error {

		if octet == 3 {
			ip, err := netip.ParseAddr(strings.Join(sipr, "."))
			result = append(result, ip)
			return err
		}

		octet++

		if strings.Contains(sipr[octet], "-") {

			rng := strings.Split(sipr[octet], "-")
			begin, err := strconv.Atoi(rng[0])
			if err != nil {
				return err
			}
			end, err := strconv.Atoi(rng[1])
			if err != nil {
				return err
			}

			for i := begin; i <= end; i++ {
				siprk := sipr
				siprk[octet] = strconv.Itoa(i)
				err := f(sipr)
				if err != nil {
					return err
				}
			}

		} else {
			err := f(sipr)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := f(splittedIP)
	if err != nil {
		return nil, err
	}

	return result, nil
}
