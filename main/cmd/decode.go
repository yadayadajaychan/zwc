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
	"io"
	"os"

	"github.com/yadayadajaychan/zwc"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode data",
	Aliases: []string{"d", "de", "dec", "deco", "decod"},

	Run: func(cmd *cobra.Command, args []string) {
		textFilename, err := cmd.Flags().GetString("text")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading text flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		if textFilename == "" || textFilename == "-" {
			textFilename = "/dev/stdin"
		}

		var text io.Reader
		if textFilename == "/dev/stdin" {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				text = bufferStdin()
			} else {
				text = os.Stdin
			}
		} else {
			text, err = os.Open(textFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc: ", err)
				os.Exit(1)
			}
		}

		decoder := zwc.NewDecoder(text)

		n, err := io.Copy(os.Stdout, decoder)
		fmt.Fprintf(os.Stderr, "zwc: %v bytes decoded\n", n)
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().StringP("text", "t", "", "Text file")

	decodeCmd.Flags().BoolP("checksum", "c", false, "Output checksum")
	decodeCmd.Flags().BoolP("message", "m", false, "Output message")

	decodeCmd.Flags().StringP("force", "f", "", "Force encoding")
}
