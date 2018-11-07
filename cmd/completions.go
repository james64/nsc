/*
 * Copyright 2018 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var bash bool
var zsh bool

var completionsCmd = &cobra.Command{
	Use:   "completions",
	Short: "Generate shell completions for bash and zsh",
	Long: `For bash completions, insure that you have bash_completion installed. 
You can do 'brew install bash-completion'. You can then enable it for the session
by doing '. /usr/local/etc/bash_completion'. For more permanent use, edit your
~/.bashrc or ~/bash_profile and add: 
'[ -f /usr/local/etc/bash_completion ] && . /usr/local/etc/bash_completion'.
Then you can simply save the output generated by 'nsc completions --bash' to a 
file and source it: 'source /path/to/savefile'.

For zsh, you'll need to find where your auto completions live - typically 
'/usr/local/share/zsh/function'. On brew installed zsh, this will be in 
'/usr/local/Cellar/zsh/5.5.1/share/zsh/functions/'. To install the completions:
'nsc completions --zsh > /usr/local/Cellar/zsh/5.5.1/share/zsh/functions/_nsc'.
Note the completion file should be called '_nsc' as shown above.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if zsh {
			rootCmd.GenZshCompletion(os.Stdout)
		}
		if bash {
			rootCmd.GenBashCompletion(os.Stdout)
		}
	},
}

func init() {
	settingsCmd.AddCommand(completionsCmd)
	completionsCmd.Flags().BoolVarP(&bash, "bash", "", false, "output bash completions")
	completionsCmd.Flags().BoolVarP(&zsh, "zsh", "", false, "output zsh completions")
}