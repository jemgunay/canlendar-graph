package config

import (
	"log"
	"os"
	"strconv"
)

// Config is the service config.
type Config struct {
	Port        int
	WebAppHost  string
	ServiceHost string
	Influx      Influx
}

// Influx contains the InfluxDB config.
type Influx struct {
	Host  string
	Token string
	Org   string
}

// New initialises a Config from environment variables.
func New() Config {
	// attempt to get config environment vars, or default them
	return Config{
		Port:        getEnvVarInt("PORT", 8080),
		WebAppHost:  getEnvVar("WEB_APP_HOST", "http://localhost:8081"),
		ServiceHost: getEnvVar("SERVICE_HOST", "http://localhost:8080"),
		Influx: Influx{
			Host:  getEnvVar("INFLUX_HOST", "http://localhost:8086"),
			Token: getEnvVar("INFLUX_TOKEN", ""),
			Org:   getEnvVar("INFLUX_ORG", ""),
		},
	}
}

// getEnvVar gets a string environment variable or defaults it if unset.
func getEnvVar(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("no %s environment var defined - defaulting to %s", key, defaultValue)
		return defaultValue
	}

	log.Printf("%s environment var found", key)
	return val
}

// getEnvVarInt gets an integer environment variable or defaults it to 0 if unset.
func getEnvVarInt(key string, defaultValue int) int {
	varStr := getEnvVar(key, strconv.Itoa(defaultValue))
	varInt, err := strconv.Atoi(varStr)
	if err != nil {
		return 0
	}
	return varInt
}
