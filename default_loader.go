package config

type DefaultLoader interface {
	LoadDefaults() error
}
