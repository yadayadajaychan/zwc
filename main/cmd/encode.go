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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/yadayadajaychan/zwc"
	"golang.org/x/term"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode data",
	Aliases: []string{"e", "en", "enc", "enco", "encod"},

	Run: func(cmd *cobra.Command, args []string) {
		dataFilename, err := cmd.Flags().GetString("data")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading data flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		messageFilename, err := cmd.Flags().GetString("message")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading message flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading interactive flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		noMessage, err := cmd.Flags().GetBool("no-message")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading no-message flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		verbose, err := cmd.Flags().GetCount("verbose")
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc: error reading verbose flag")
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		//quiet, err := cmd.Flags().GetBool("quiet")
		//if err != nil {
		//	fmt.Fprintln(os.Stderr, "zwc: error reading quiet flag")
		//	fmt.Fprintln(os.Stderr, "zwc:", err)
		//	os.Exit(2)
		//}

		// interactive has no effect if data or message are supplied
		// or if no-message is specified
		if dataFilename != "" || messageFilename != "" || noMessage {
			interactive = false
		}

		// if message is supplied, no-message has no effect
		if messageFilename != "" {
			noMessage = false
		}

		if noMessage && (dataFilename == "" || dataFilename == "-") {
			dataFilename = "/dev/stdin"
		} else if dataFilename == "" && messageFilename == "" && !interactive {
			fmt.Fprintln(os.Stderr, "zwc: data and/or message file must be specified")
			os.Exit(1)
		} else if dataFilename == messageFilename && !interactive {
			fmt.Fprintln(os.Stderr, "zwc: data and message file must be different")
			os.Exit(1)
		} else if dataFilename == "" || dataFilename == "-" {
			dataFilename = "/dev/stdin"
		} else if messageFilename == "" || messageFilename == "-" {
			messageFilename = "/dev/stdin"
		}

		encoding := createEncoding(cmd)
		encoder := zwc.NewEncoder(encoding, os.Stdout)

		var data, message io.Reader

		if interactive {
			var dataBuffer, messageBuffer bytes.Buffer
			fmt.Fprintln(os.Stderr, "Enter data, then the string 'EOF' on its own line, then the message, then Ctrl-D.")

			// whether or not the string "EOF" has been seen
			eof := false
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				str := scanner.Text()
				if eof {
					messageBuffer.WriteString(str + "\n")
				} else if str != "EOF" {
					dataBuffer.WriteString(str + "\n")
				} else {
					eof = true
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "zwc:", err)
				os.Exit(2)
			}

			if dataBuffer.Len() == 0 {
				fmt.Fprintln(os.Stderr, "zwc: no data supplied")
				os.Exit(1)
			}

			if messageBuffer.Len() == 0 {
				fmt.Fprintln(os.Stderr, "zwc: no message supplied")
				os.Exit(1)
			}

			data = &dataBuffer
			message = &messageBuffer
		} else if noMessage {
			if dataFilename == "/dev/stdin" {
				if term.IsTerminal(int(os.Stdin.Fd())) {
					data = bufferStdin()
				} else {
					data = os.Stdin
				}
			} else {
				data, err = os.Open(dataFilename)
				if err != nil {
					fmt.Fprintln(os.Stderr, "zwc:", err)
					os.Exit(1)
				}
			}
		} else if dataFilename == "/dev/stdin" && term.IsTerminal(int(os.Stdin.Fd())) {
			// buffer data if connected to terminal
			if verbose >= 1 {
				fmt.Fprintln(os.Stderr, "zwc: reading data from terminal")
			}

			data = bufferStdin()

			message, err = os.Open(messageFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc:", err)
				os.Exit(1)
			}
		} else if messageFilename == "/dev/stdin" && term.IsTerminal(int(os.Stdin.Fd())) {
			//buffer message if connected to a terminal
			if verbose >= 1 {
				fmt.Fprintln(os.Stderr, "zwc: reading message from terminal")
			}

			message = bufferStdin()

			data, err = os.Open(dataFilename)
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc:", err)
				os.Exit(1)
			}
		} else {
			if dataFilename == "/dev/stdin" {
				// /dev/stdin may not exist on all systems
				data = os.Stdin
			} else {
				data, err = os.Open(dataFilename)
				if err != nil {
					fmt.Fprintln(os.Stderr, "zwc:", err)
					os.Exit(1)
				}
			}

			if messageFilename == "/dev/stdin" {
				// /dev/stdin may not exist on all systems
				message = os.Stdin
			} else {
				message, err = os.Open(messageFilename)
				if err != nil {
					fmt.Fprintln(os.Stderr, "zwc:", err)
					os.Exit(1)
				}
			}
		}


		var fmi int
		if !noMessage {
			// fm holds first character from message
			fm := make([]byte, utf8.UTFMax)
			fmi, err = message.Read(fm[:1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc:", err)
				os.Exit(2)
			}

			// read more bytes if not full character
			for !utf8.FullRune(fm[:fmi]) {
				n, err := message.Read(fm[fmi:fmi+1])
				fmi += n
				if err != nil {
					fmt.Fprintln(os.Stderr, "zwc:", err)
					os.Exit(2)
				}
			}

			// write first character from message
			if _, err := os.Stdout.Write(fm[:fmi]); err != nil {
				fmt.Fprintln(os.Stderr, "zwc:", err)
				os.Exit(2)
			}
		}

		// encode data
		nDataEncoded, err := io.Copy(encoder, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		err = encoder.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "zwc:", err)
			os.Exit(2)
		}

		if !noMessage {
			// write rest of message
			n, err := io.Copy(os.Stdout, message)
			if verbose >= 3 {
				fmt.Fprintf(os.Stderr, "zwc: %v bytes of data encoded\n", n + int64(fmi))
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "zwc:", err)
				os.Exit(2)
			}
		}

		if verbose >= 2 {
			fmt.Fprintf(os.Stderr, "zwc: version %v, encoding %v, checksum %v\n",
						encoding.Version(), encoding.EncodingType(), encoding.ChecksumType())
			fmt.Fprintf(os.Stderr, "zwc: %v bytes of data encoded\n", nDataEncoded)
			fmt.Fprintf(os.Stderr, "zwc: crc is %x\n", encoding.Checksum())
		}
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)

	encodeCmd.Flags().StringP("data", "d", "", "Data file")
	encodeCmd.Flags().StringP("message", "m", "", "Message file")

	encodeCmd.Flags().IntP("checksum", "c", 16, "Checksum type")
	encodeCmd.Flags().IntP("encoding", "e", 3, "Encoding type")

	encodeCmd.Flags().BoolP("interactive", "i", false, "Interactive mode")
	encodeCmd.Flags().BoolP("no-message", "n", false, "No message")
}

func createEncoding(cmd *cobra.Command) *zwc.Encoding {
	checksum, err := cmd.Flags().GetInt("checksum")
	if err != nil {
		fmt.Fprintln(os.Stderr, "zwc: error reading checksum flag")
		fmt.Fprintln(os.Stderr, "zwc:", err)
		os.Exit(2)
	}

	encoding, err := cmd.Flags().GetInt("encoding")
	if err != nil {
		fmt.Fprintln(os.Stderr, "zwc: error reading encoding flag")
		fmt.Fprintln(os.Stderr, "zwc:", err)
		os.Exit(2)
	}

	switch checksum {
	case 0, 8, 16, 32:
	default:
		fmt.Fprintln(os.Stderr, "zwc: invalid checksum type of", checksum)
		fmt.Fprintln(os.Stderr, "zwc: checksum must be either 0, 8, 16, or 32")
		os.Exit(1)
	}

	switch encoding {
	case 2, 3, 4:
	default:
		fmt.Fprintln(os.Stderr, "zwc: invalid encoding type of", encoding)
		fmt.Fprintln(os.Stderr, "zwc: encoding must be either 2, 3, or 4")
		os.Exit(1)
	}

	return zwc.NewEncoding(1, encoding, checksum)
}

func bufferStdin() *bytes.Buffer {
	var buffer bytes.Buffer

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "zwc:", err)
		os.Exit(2)
	}

	return &buffer
}
