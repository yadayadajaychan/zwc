# Specification Version 0.2 (Draft)

## Data encoding

| data  | unicode |        description        |  utf-8   |
|-------|---------|---------------------------|----------|
| 0     | U+200B  | zero-width space          | 0xE2808B |
| 1     | U+200C  | zero width non-joiner     | 0xE2808C |
| 2     | U+200D  | zero-width joiner         | 0xE2808D |
| 3     | U+2060  | word-joiner               | 0xE281A0 |
| delim | U+034F  | combining grapheme joiner | 0xCD8F   |

Each character encodes two bits of data. There is also a delimiter character
to delimit the header, payload, and checksum from each other. Bytes of data are
encoded with the most significant bits first.

## Layout

| *file signature* | *header* | delim | *payload* | delim | *checksum* | delim |
|------------------|----------|-------|-----------|-------|------------|-------|

This data is interspersed among a message with non-zero-width characters. The
message must not contain any of the zero-width characters used to encode the data.

## File signature

0 followed by delim (U+200B U+034F). Must appear directly after the first non-whitespace
character.

## Header

The header appears after the file signature and is a 3 bit field with 1 parity
bit (4 bits total). The parity bit is the most significant bit and the parity is
even. Below are the possible configurations:

| num |      config      |
|-----|------------------|
|   0 | none             |
|   1 | crc-8            |
|   2 | crc-16           |
|   3 | crc-32           |
|   4 | crc-64           |
|   5 | md5              |
|   6 | sha-256          |
|   7 | Reed-Solomon ECC |

E.g. to set the checksum as crc-64, the header would be
0b1100 -> 3 0 -> U+2060 U+200B

## Payload

The actual data being hidden by the user is encoded in the payload. Each byte
will require 4 zero-width characters to encode it. E.g. 0b10110100 -> 2 3 1 0 ->
U+200D U+2060 U+200C U+200B. The payload is separated from the header and
checksum by the delim character.

If Reed-Solomon ECC is specified, the data is encoded with ECC before being
converted to zero-width characters.

## Checksum

This section contains the encoded checksum and it must be ended with a delim
character even if there is no checksum.
