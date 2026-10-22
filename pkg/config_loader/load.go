package config_loader

import (
	"fmt"
	"os"
	"sync"

	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/caarlos0/env"
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	validate = validator.New()
)

var configs = []configLoader{}

func Register(name string, cfg Config) {
	configs = append(configs, NewConfig(name, cfg)) // cfg, not &cfg
}

func LoadAndValidate(dot_env_validation bool) {
	once.Do(func() {
		env_file := os.Getenv("ENV")
		if env_file == "" {
			env_file = ".env"
		}

		if err := godotenv.Load(env_file); err != nil {
			simple_logger.Warnf("configs.Init could not load .env error: %v", err)
			simple_logger.Fatalf("%v", err)
		}

		for _, cfg := range configs {
			if err := env.Parse(cfg.Ptr()); err != nil {
				simple_logger.Fatalf("%s: %v", cfg.Name(), err)
			}
		}

		if dot_env_validation {
			validateConfig()
		}

	})
}

func validateConfig() {
	var errs []string

	for _, cfg := range configs {
		if err := validate.Struct(cfg.Ptr()); err != nil {
			errs = append(errs, fmt.Sprintf("invalid %s: %v", cfg.Name(), err))
			continue
		}

		v, ok := cfg.Ptr().(Config)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid config %s", cfg.Name()))
			continue
		}

		if err := v.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", cfg.Name(), err))
		}
	}

	if len(errs) == 0 {
		return
	}

	for _, e := range errs {
		simple_logger.Errorf("%s", e)
	}
	simple_logger.Fatalf("configuration validation failed with %d error(s)", len(errs))
}
