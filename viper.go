package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func setViperFlags(f *pflag.FlagSet, cfg interface{}) {
	typ := reflect.TypeOf(cfg).Elem()
	val := reflect.ValueOf(cfg).Elem()
	var err error
BAD:
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name, nameExists := field.Tag.Lookup("mapstructure")
		if !nameExists {
			continue
		}
		keys := strings.Split(name, ",")
		if len(keys) > 1 {
			if strings.Contains(name, "omitempty") {
				continue
			}
			name = keys[0]
		}

		var defaultValue interface{}
		isTypePtr := field.Type.Kind() == reflect.Ptr
		defaultValueString, _ := field.Tag.Lookup("default")
		if defaultValueString != "" {
			switch field.Type.String() {
			case "*bool", "bool":
				defaultValue, err = strconv.ParseBool(defaultValueString)
			case "*int", "int":
				defaultValue, err = strconv.Atoi(defaultValueString)
			case "*int64", "int64":
				defaultValue, err = strconv.ParseInt(defaultValueString, 10, 64)
			case "*uint", "uint":
				var u uint64
				u, err = strconv.ParseUint(defaultValueString, 10, 64)
				defaultValue = uint(u)
			case "*uint64", "uint64":
				defaultValue, err = strconv.ParseUint(defaultValueString, 10, 64)
			case "*float64", "float64":
				defaultValue, err = strconv.ParseFloat(defaultValueString, 64)
			case "*string", "string":
				defaultValue = defaultValueString
			case "*time.Duration", "time.Duration":
				defaultValue, err = time.ParseDuration(defaultValueString)
			default:
				err = fmt.Errorf("flag default value type '%s' not support", field.Type.String())
			}
			if err != nil {
				err = fmt.Errorf("flag parse error '%s' :%w", defaultValueString, err)
				break BAD
			}
		}

		el := val.Field(i)
		if el.Kind() == reflect.Ptr {
			if !el.IsNil() {
				defaultValue = el.Elem().Interface()
			}
		} else if !el.IsZero() {
			defaultValue = el.Interface()
		}

		if defaultValue == nil {
			if isTypePtr {
				defaultValue = reflect.Zero(field.Type.Elem()).Interface()
			} else {
				defaultValue = reflect.Zero(field.Type).Interface()
			}
		}

		description, _ := field.Tag.Lookup("desc")
		shorthand, _ := field.Tag.Lookup("short")
		switch field.Type.String() {
		case "*bool", "bool":
			f.BoolP(name, shorthand, defaultValue.(bool), description)
		case "*int", "int":
			f.IntP(name, shorthand, defaultValue.(int), description)
		case "*int64", "int64":
			f.Int64P(name, shorthand, defaultValue.(int64), description)
		case "*uint", "uint":
			f.UintP(name, shorthand, defaultValue.(uint), description)
		case "*uint64", "uint64":
			f.Uint64P(name, shorthand, defaultValue.(uint64), description)
		case "*float64", "float64":
			f.Float64P(name, shorthand, defaultValue.(float64), description)
		case "*string", "string":
			f.StringP(name, shorthand, defaultValue.(string), description)
		case "*time.Duration", "time.Duration":
			f.DurationP(name, shorthand, defaultValue.(time.Duration), description)
		default:
			err = fmt.Errorf("flag type '%s' not support", field.Type.String())
			break BAD
		}
	}
	if err != nil {
		panic(err)
	}
}

func init() {
	cf := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	cf.SetOutput(ioutil.Discard)

	configPath := cf.StringP("config", "c", filepath.Join(DefaultConfigPath, DefaultConfigName), "path to configuration")
	_ = cf.Parse(os.Args[1:])

	f := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	_ = f.StringP("config", "c", filepath.Join(DefaultConfigPath, DefaultConfigName), "path to configuration")
	showVersion := f.BoolP("version", "v", false, "show version")
	setViperFlags(f, &Config)
	_ = f.Parse(os.Args[1:])

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()

	if *configPath != "" {
		if !strings.HasPrefix(*configPath, "/") {
			*configPath = "./" + *configPath
		}
		ext := filepath.Ext(*configPath)
		*configPath = strings.TrimSuffix(*configPath, ext)
		base := filepath.Base(*configPath)
		v.SetConfigName(base)
		v.AddConfigPath(strings.TrimSuffix(*configPath, base))
	} else {
		v.SetConfigName(DefaultConfigName)
		v.AddConfigPath(DefaultConfigPath)
	}

	_ = v.ReadInConfig()
	if err := v.BindPFlags(f); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&Config); err != nil {
		panic(err)
	}
}
