// Copyright (C) 2023 Ethan Cheng <ethanrc0528@gmail.com>
//
// This file is part of ZWC.
//
// ZWC is free software: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, version 3 of the License.
//
// ZWC is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along
// with ZWC. If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version info",
	Long: "Display program version, file format specification version, and copyright info.",
	Aliases: []string{"v", "ve", "ver", "vers", "versi", "versio"},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ZWC Program Version %v\n", version)
		fmt.Printf("ZWC File Format Version %v\n\n", fileFormat)
		fmt.Println(`Copyright (C) 2023 Ethan Cheng <ethanrc0528@gmail.com>
License GPLv3: GNU GPL version 3 <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
