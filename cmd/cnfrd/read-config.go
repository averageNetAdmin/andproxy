package cnfrd

import (
	"errors"
	"log"
	"strings"

	"github.com/averageNetAdmin/andproxy/cmd/ipdb"
	"github.com/averageNetAdmin/andproxy/cmd/porthndlr"
	"github.com/spf13/viper"
)

func ReadConfig(CONFIGFILEPATH string) ([]*porthndlr.PortHandler, error) {

	viper.SetConfigType("yaml")
	viper.SetConfigFile(CONFIGFILEPATH)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var logDir string
	if viper.IsSet("global.logDir") {
		a, ok := (viper.Get("global.logDir")).(string)
		if !ok {
			return nil, errors.New("invalid log directory")
		}
		logDir = a
	} else {
		logDir = "/var/log"
	}

	db := ipdb.NewIPDB()
	if viper.IsSet("pools") {
		a, ok := (viper.Get("pools")).(map[string]interface{})
		if !ok {
			return nil, errors.New("error during reading address pools stage 1")
		}
		for name, p := range a {
			pool, ok := p.([]interface{})
			if !ok {
				return nil, errors.New("error during reading address pools stage 2")
			}
			err := db.AddPool(name, pool)
			if err != nil {
				return nil, err
			}
		}

	}

	if viper.IsSet("serversPools") {
		a, ok := (viper.Get("serversPools")).(map[string]interface{})
		log.Println(a)
		if !ok {
			return nil, errors.New("error during reading servers pools")
		}
		for name, p := range a {
			pool, ok := p.(map[string]interface{})
			if !ok {
				return nil, errors.New("error during reading servers pools")
			}
			err := db.AddServerPool(name, pool)
			if err != nil {
				return nil, err
			}
		}

	}

	if viper.IsSet("filters") {
		a, ok := (viper.Get("filters")).(map[string]interface{})
		if !ok {
			return nil, errors.New("error during reading filters")
		}
		for name, p := range a {
			pool, ok := p.(map[string]interface{})
			if !ok {
				return nil, errors.New("error during reading filters")
			}
			err := db.AddFilter(name, pool)
			if err != nil {
				return nil, err
			}
		}
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
			hndlr, err := porthndlr.NewPortHandler(nameparts[0], nameparts[1], db, conf, logDir+"/andproxy")
			if err != nil {
				return nil, err
			}
			hndlrs = append(hndlrs, hndlr)
		}
	}
	return hndlrs, nil
}
