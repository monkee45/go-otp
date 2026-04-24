package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Database PostgresConfig `json:"database"`
}

func (c PostgresConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Name)

	} else {
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Name)
	}
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Database: DefaultPostgreConfig(),
	}
}

// parameter values needed to open/connect to the database
func DefaultPostgreConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "michaelwalsh",
		Password: "mondo",
		Name:     "story_dev",
	}
}

// LoadConfig allows for the use of '.config' file to hold config details
func LoadConfig(configReq bool) Config {
	//open json file from current directory
	f, err := os.Open(".config")
	if err != nil {
		if configReq {
			panic(err)
		}
		fmt.Println("Using the default config ....")
		return DefaultConfig()
	}
	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully loaded .config")
	return c
}
