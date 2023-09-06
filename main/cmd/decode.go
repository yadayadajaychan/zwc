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
	"strconv"
	"strings"

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

		force, err := cmd.Flags().GetString("force")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading force flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading quiet flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		verbose, err := cmd.Flags().GetCount("verbose")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading verbose flag")
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

		var decoder io.Reader
		var encoding *zwc.Encoding
		var v, e, c int

		if force == "" {
			v, e, c, err = zwc.DecodeHeaderFromReader(text)
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc: ", err)
				os.Exit(2)
			}

			encoding = zwc.NewEncoding(v, e, c)
			decoder = zwc.NewCustomDecoder(encoding, text)
		} else {
			v, e, c = parseForce(force)

			// ignore values from header
			_, _, _, err = zwc.DecodeHeaderFromReader(text)
			if err != nil && !quiet {
				fmt.Fprintln(os.Stderr, "zwc: warning: ", err)
			}

			encoding = zwc.NewEncoding(v, e, c)
			decoder = zwc.NewCustomDecoder(encoding, text)
		}

		n, err := io.Copy(os.Stdout, decoder)
		if verbose >= 2 {
			fmt.Fprintf(os.Stderr, "zwc: version %v, encoding %v, checksum %v\n", v, e, c)
			fmt.Fprintf(os.Stderr, "zwc: %v bytes decoded\n", n)
			fmt.Fprintf(os.Stderr, "zwc: crc is %x\n", encoding.Checksum())
		}
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

// parse force flag
func parseForce(force string) (v, e, c int) {
	f := strings.Split(force, ",")

	// check number of values
	if len(f) != 2 {
		fmt.Fprintln(os.Stderr, "zwc: force flag requires two comma-separated values for encoding and checksum type")
		os.Exit(1)
	}

	// convert to int
	var fconv [2]int
	var err error

	fconv[0], err = strconv.Atoi(f[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "zwc: force argument contains non-integer value(s)")
		os.Exit(1)
	}

	fconv[1], err = strconv.Atoi(f[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "zwc: force argument contains non-integer value(s)")
		os.Exit(1)
	}

	switch fconv[0] {
	case 2, 3, 4:
		e = fconv[0]
	case 0, 8, 16, 32:
		c = fconv[0]
	default:
		fmt.Fprintln(os.Stderr, "zwc: force argument contains invalid encoding/checksum type")
		os.Exit(1)
	}

	switch fconv[1] {
	case 2, 3, 4:
		e = fconv[1]
	case 0, 8, 16, 32:
		c = fconv[1]
	default:
		fmt.Fprintln(os.Stderr, "zwc: force argument contains invalid encoding/checksum type")
		os.Exit(1)
	}

	v = 1
	return v, e, c
}
