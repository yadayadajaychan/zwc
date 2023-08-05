# ZWC File Format Specification Version 0.9 (Draft)

The ZWC format describes how data should be encoded as zero-width characters.
This encoded data is then put inside a message of non-zero-width characters.
The result is a text stream with a visible message but hidden data.

## Data encoding

### 2-bit encoding

| data  | unicode |        description         |   utf-8    |
|-------|---------|----------------------------|------------|
| delim | U+034F  | combining grapheme joiner  | 0xCD 8F    |
| 0     | U+202C  | pop directional formatting | 0xE2 80 AC |
| 1     | U+200C  | zero width non-joiner      | 0xE2 80 8C |
| 2     | U+200D  | zero-width joiner          | 0xE2 80 8D |
| 3     | U+2060  | word-joiner                | 0xE2 81 A0 |

Each character encodes two bits of data. There is also a delimiter character
to delimit the header, payload, and checksum from each other. Bytes of data are
split into four and encoded with the most significant bits first.  
E.g. 0b10110100 -> 2 3 1 0 -> U+200D U+2060 U+200C U+202C.

### 3-bit encoding

Extension of the 2 bit encoding.

| data | unicode |     description      |   utf-8    |
|------|---------|----------------------|------------|
|    4 | U+2061  | function application | 0xE2 81 A1 |
|    5 | U+2062  | invisible times      | 0xE2 81 A2 |
|    6 | U+2063  | invisible separator  | 0xE2 81 A3 |
|    7 | U+2064  | invisible plus       | 0xE2 81 A4 |

Each character encodes three bits of data. The process is the same as the two
bit encoding but now bytes are split into three.  
E.g. 0b10110100 -> 2 6 4 -> U+200D U+2063 U+2061.

### 4-bit encoding

Extension of 3 bit encoding.

| data | unicode |         description          |     utf-8     |
|------|---------|------------------------------|---------------|
|    8 | U+206A  | inhibit symmetric swapping   | 0xE2 81 AA    |
|    9 | U+206B  | activate symmetric swapping  | 0xE2 81 AB    |
|   10 | U+206C  | inhibit arabic form shaping  | 0xE2 81 AC    |
|   11 | U+206D  | activate arabic form shaping | 0xE2 81 AD    |
|   12 | U+206E  | national digit shapes        | 0xE2 81 AE    |
|   13 | U+206F  | nominal digit shapes         | 0xE2 81 AF    |
|   14 | U+1D173 | musical symbol begin beam    | 0xF0 9D 85 B3 |
|   15 | U+1D174 | musical symbol end beam      | 0xF0 9D 85 B4 |

Each character encodes four bits of data. Bytes are split into two.  
E.g. 0b10110100 -> 11 4 -> U+206D U+2061.

## Layout

| *file signature* | *header* | delim | *payload* | delim | *checksum* |
|------------------|----------|-------|-----------|-------|------------|

The encoded data is interspersed among a message with non-zero-width characters.
When decoding, any characters not in the data encoding table for the selected
encoding type are ignored. The message must not contain any of the zero-width
characters used to encode the data. There may be multiple files within the same
message.

## File signature

delim (U+034F)

## Header

The header uses 2-bit encoding regardless of what encoding is used for the
payload and checksum. The header appears after the file signature and is 8 bits
long.

| Field Name | Offset (bits) | Length (bits) |           Description            |
|------------|---------------|---------------|----------------------------------|
| version    |             0 |             2 | major version of zwc file format |
| encoding   |             2 |             2 | encoding used for the payload    |
| checksum   |             4 |             2 | checksum used for the payload    |
| crc-2      |             6 |             2 | crc used to protect the header   |

Below are the possible configurations:

| version | value |
|---------|-------|
| v1      |     0 |
| v2      |     1 |
| v3      |     2 |
| v4      |     3 |

| encoding | value |
|----------|-------|
| 2-bit    |     0 |
| 3-bit    |     1 |
| 4-bit    |     2 |

|     checksum     | value |
|------------------|-------|
| none             |     0 |
| crc-8            |     1 |
| crc-16           |     2 |
| crc-32           |     3 |

E.g. to set the file format as version 2, the encoding as 4-bit and the checksum
as crc-32, the header would be 0b01\_10\_11_(crc-2).

### CRC-2

Below are the parameters for the crc used to protect the header:

WIDTH: 2  
POLY: 0x03  
INIT: 0x00  
REFIN: FALSE  
REFOUT: FALSE  
XOROUT: 0x00

## Payload

The actual data being hidden by the user is encoded in the payload. Each byte
will require 4 to 2 zero-width characters to encode it, depending on the
specified encoding. The payload is separated from the header and checksum by
the delim character.

## Checksum

This section contains the encoded checksum and must not end with a delim
character. The checksum uses the same encoding as the payload.

### CRC-8-CCITT

WIDTH: 8  
POLY: 0x07  
INIT: 0x00  
REFIN: FALSE  
REFOUT: FALSE  
XOROUT: 0x00  
CHECK: 0xF4  

### CRC-16-CCITT

WIDTH: 16  
POLY: 0x1021  
INIT: 0x0000  
REFIN: FALSE  
REFOUT: FALSE  
XOROUT: 0x0000  
CHECK: 0x31C3  

### CRC-32

WIDTH: 32  
POLY: 0x04C11DB7  
INIT: 0xFFFFFFFF  
REFIN: TRUE  
REFOUT: TRUE  
XOROUT: 0xFFFFFFFF  
CHECK: 0xCBF43926  

# Copyright Information

Copyright (C) 2023 Ethan Cheng

This file is part of ZWC.

ZWC is free software: you can redistribute it and/or modify it under the
terms of the GNU General Public License as published by the Free Software
Foundation, version 3 of the License.

ZWC is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
details.

You should have received a copy of the GNU General Public License along
with ZWC. If not, see <https://www.gnu.org/licenses/>.
