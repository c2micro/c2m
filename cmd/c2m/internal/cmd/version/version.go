package version

import (
	"fmt"

	"github.com/c2micro/c2m/internal/version"
	"github.com/spf13/cobra"
)

type cmd struct{}

func (cmd) Run(*cobra.Command, []string) error {
	fmt.Println(version.Get().PrettyColorful())
	return nil
}
