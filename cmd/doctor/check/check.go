package check

import "github.com/candrewlee14/webman/config"

type Check struct {
	Name string
	Func func(cfg *config.Config, fix bool) error
}
