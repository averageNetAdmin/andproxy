package config

import (
	"fmt"
	"strings"

	"github.com/averageNetAdmin/andproxy/cmd/ipdb"
	"github.com/averageNetAdmin/andproxy/cmd/porthndlr"
	"github.com/spf13/viper"
)

func ReadAndCreate(CONFIGFILEPATH string) (map[string]*porthndlr.Handler, string, error) {

	a, db, logDir, err := read(CONFIGFILEPATH)
	if err != nil {
		return nil, "", err
	}

	hdnlrs := make(map[string]*porthndlr.Handler)

	for name, p := range a {
		if p == nil {
			continue
		}
		c, ok := p.(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("error during reading listenning ports")
		}
		nameparts := strings.Split(name, " ")
		hndlr, err := porthndlr.NewHandler(nameparts[0], nameparts[1], db, c, fmt.Sprintf("%s/andproxy", logDir))
		if err != nil {
			return nil, "", err
		}
		hdnlrs[name] = hndlr
	}
	return hdnlrs, logDir, nil
}

func Read(CONFIGFILEPATH string) (map[string]*porthndlr.Config, error) {

	a, db, logDir, err := read(CONFIGFILEPATH)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]*porthndlr.Config)

	for name, p := range a {
		if p == nil {
			continue
		}
		c, ok := p.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error during reading listenning ports")
		}
		nameparts := strings.Split(name, " ")
		conf, err := porthndlr.NewConfig(c, db, nameparts[1], fmt.Sprintf("%s/andproxy", logDir))
		if err != nil {
			return nil, err
		}
		configs[name] = conf
	}
	return configs, nil
}

func read(CONFIGFILEPATH string) (map[string]interface{}, *ipdb.IPDB, string, error) {

	viper.SetConfigType("yaml")
	viper.SetConfigFile(CONFIGFILEPATH)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, nil, "", err
	}
	var logDir string
	if viper.IsSet("global.logDir") {
		a, ok := (viper.Get("global.logDir")).(string)
		if !ok {
			return nil, nil, "", fmt.Errorf("invalid log directory")
		}
		logDir = a
	} else {
		logDir = "/var/log"
	}

	db := ipdb.NewIPDB()
	if viper.IsSet("pools") {
		a, ok := (viper.Get("pools")).(map[string]interface{})
		if !ok {
			return nil, nil, "", fmt.Errorf("error during reading address pools stage 1")
		}
		for name, p := range a {
			pool, ok := p.([]interface{})
			if !ok {
				return nil, nil, "", fmt.Errorf("error during reading address pools stage 2")
			}
			err := db.AddPool(name, pool)
			if err != nil {
				return nil, nil, "", err
			}
		}

	}

	if viper.IsSet("serversPools") {
		a, ok := (viper.Get("serversPools")).(map[string]interface{})
		if !ok {
			return nil, nil, "", fmt.Errorf("error during reading servers pools")
		}
		for name, p := range a {
			pool, ok := p.(map[string]interface{})
			if !ok {
				return nil, nil, "", fmt.Errorf("error during reading servers pools")
			}
			err := db.AddServerPool(name, pool)
			if err != nil {
				return nil, nil, "", err
			}
		}

	}

	if viper.IsSet("filters") {
		a, ok := (viper.Get("filters")).(map[string]interface{})
		if !ok {
			return nil, nil, "", fmt.Errorf("error during reading filters")
		}
		for name, p := range a {
			pool, ok := p.(map[string]interface{})
			if !ok {
				return nil, nil, "", fmt.Errorf("error during reading filters")
			}
			err := db.AddFilter(name, pool)
			if err != nil {
				return nil, nil, "", err
			}
		}
	}
	res := make(map[string]interface{})
	if viper.IsSet("listenPorts") {
		a, ok := (viper.Get("listenPorts")).(map[string]interface{})
		if !ok {
			return nil, nil, "", fmt.Errorf("error during reading listenning ports")
		}
		res = a
	}
	return res, db, logDir, nil
}
