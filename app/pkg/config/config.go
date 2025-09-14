package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	mu    sync.RWMutex
	store map[string]any
}

func NewConfig() *Config {
	return &Config{
		store: make(map[string]any),
	}
}

func (c *Config) LoadEnvVar(key, defaultValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if envValue, exists := os.LookupEnv(key); exists {
		c.store[key] = envValue
	} else {
		c.store[key] = defaultValue
	}
}

func (c *Config) LoadEnv() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		c.store[parts[0]] = parts[1]
	}
}

func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
}

func (c *Config) GetString(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.store[key]
	if !exists {
		return "", errors.New("config key not found: " + key)
	}
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", errors.New("config key is not a string: " + key)
	}
}

func (c *Config) GetInt(key string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.store[key]
	if !exists {
		return 0, errors.New("config key not found: " + key)
	}
	switch v := value.(type) {
	case int:
		return v, nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, errors.New("config key cannot be parsed as int: " + key)
		}
		return i, nil
	default:
		return 0, errors.New("config key is not an int: " + key)
	}
}

func (c *Config) GetBool(key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.store[key]
	if !exists {
		return false, errors.New("config key not found: " + key)
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return false, errors.New("config key cannot be parsed as bool: " + key)
		}
		return b, nil
	default:
		return false, errors.New("config key is not a bool: " + key)
	}
}

func (c *Config) GetFloat(key string) (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.store[key]
	if !exists {
		return 0, errors.New("config key not found: " + key)
	}
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, errors.New("config key cannot be parsed as float: " + key)
		}
		return f, nil
	default:
		return 0, errors.New("config key is not a float: " + key)
	}
}

// MustString returns the string value or panics if missing/invalid
func (c *Config) MustString(key string) string {
	val, err := c.GetString(key)
	if err != nil {
		panic(fmt.Sprintf("missing or invalid config key %q: %v", key, err))
	}
	return val
}

// MustInt returns the int value or panics if missing/invalid
func (c *Config) MustInt(key string) int {
	val, err := c.GetInt(key)
	if err != nil {
		panic(fmt.Sprintf("missing or invalid config key %q: %v", key, err))
	}
	return val
}

// MustBool returns the bool value or panics if missing/invalid
func (c *Config) MustBool(key string) bool {
	val, err := c.GetBool(key)
	if err != nil {
		panic(fmt.Sprintf("missing or invalid config key %q: %v", key, err))
	}
	return val
}

// MustFloat returns the float64 value or panics if missing/invalid
func (c *Config) MustFloat(key string) float64 {
	val, err := c.GetFloat(key)
	if err != nil {
		panic(fmt.Sprintf("missing or invalid config key %q: %v", key, err))
	}
	return val
}
