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
		Short: "Generate completion script",
		Long: `To load completions:
Bash:

$ source <(./bin/chaosd completion bash)

# To load completions for each session, execute once:
Linux:
	$ ./bin/chaosd completion bash > /etc/bash_completion.d/chaosd
MacOS:
	$ ./bin/chaosd completion bash > /usr/local/etc/bash_completion.d/chaosd

Zsh:

$ compdef _chaosd ./bin/chaosd

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ ./bin/chaosd completion zsh > "${fpath[1]}/_chaosd"

# You will need to start a new shell for this setup to take effect.

Fish:

$ ./bin/chaosd completion fish | source

# To load completions for each session, execute once:
$ ./bin/chaosd completion fish > ~/.config/fish/completions/chaosd.fish

Powershell:

PS> ./bin/chaosd completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> ./bin/chaosd completion powershell > chaosd.ps1
# and source this file from your powershell profile.
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
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
