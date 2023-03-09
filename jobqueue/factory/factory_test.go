package factory

import (
	"testing"

	"github.com/coosir/middleman/config"
	"github.com/coosir/middleman/model"
)

func TestInvalidDriver(t *testing.T) {
	config.Locally("driver", "nothing", func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("It should die")
			}
		}()

		NewImpl(&model.Queue{})
	})
}
