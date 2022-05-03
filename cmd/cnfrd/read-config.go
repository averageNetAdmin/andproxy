package cnfrd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/averageNetAdmin/andproxy/source/ipdb"
	"github.com/averageNetAdmin/andproxy/source/porthndlr"
	"github.com/spf13/viper"
)

func ReadConfig(CONFIGFILEPATH string) ([]*porthndlr.PortHandler, error) {

	viper.SetConfigType("yaml")
	viper.SetConfigFile(CONFIGFILEPATH)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	db := new(ipdb.IPDB)
	if viper.IsSet("addressPools") {
		a, ok := (viper.Get("addressPools")).(map[string]interface{})
		if !ok {
			return nil, errors.New("error during reading address pools")
		}
		for name, p := range a {
			pool, ok := p.([]interface{})
			if !ok {
				return nil, errors.New("error during reading address pools")
			}
			db.AddPool(name, pool)
		}

	}

	if viper.IsSet("servers") {
		a, ok := (viper.Get("servers")).(map[string]interface{})
		if !ok {
			return nil, errors.New("error during reading servers pools")
		}
		for name, p := range a {
			pool, ok := p.([]interface{})
			if !ok {
				return nil, errors.New("error during reading servers pools")
			}
			db.AddServerPool(name, pool)
		}
		fmt.Println(db)

	}

	if viper.IsSet("staticFilters") {
		a, ok := (viper.Get("staticFilters")).(map[string]interface{})
		fmt.Println(a)
		if !ok {
			return nil, errors.New("error during reading filters")
		}
		for name, p := range a {
			pool, ok := p.(map[string]interface{})
			fmt.Println(name, pool)
			if !ok {
				return nil, errors.New("error during reading filters")
			}
			db.AddFilter(name, pool)
		}
		fmt.Println(db)
	}
	hndlrs := make([]*porthndlr.PortHandler, 0)
	if viper.IsSet("listenPorts") {
		a, ok := (viper.Get("listenPorts")).(map[string]interface{})
		if !ok {
			return nil, errors.New("error during reading listenning ports")
		}
		for name, p := range a {
			nameparts := strings.Split(name, " ")
			if p == nil {
				continue
			}
			conf, ok := p.(map[string]interface{})
			if !ok {
				return nil, errors.New("error during reading listenning ports")
			}
			hndlr, err := porthndlr.NewPortHandler(nameparts[0], nameparts[1], db, conf)
			if err != nil {
				return nil, err
			}
			hndlrs = append(hndlrs, hndlr)
		}
	}
	return hndlrs, nil
}
