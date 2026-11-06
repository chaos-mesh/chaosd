// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package completion

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

// NewCompletionCommand returns the completion command
func NewCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate a shell completion script",
		Long: `Generate a completion script for Bash, Zsh, or Fish.

Load completions in the current shell:

Bash:
  source <(chaosd completion bash)

Zsh:
  autoload -U compinit && compinit
  source <(chaosd completion zsh)

Fish:
  chaosd completion fish | source

To load completions in future sessions, add the appropriate command to your
shell's startup file.

PowerShell is accepted as an argument but currently produces no script.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
					utils.ExitWithError(utils.ExitError, err)
				}
			case "zsh":
				if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
					utils.ExitWithError(utils.ExitError, err)
				}
			case "fish":
				if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
					utils.ExitWithError(utils.ExitError, err)
				}
			}
		},
	}
}
