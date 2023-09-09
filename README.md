# ZWC

zwc is a program for encoding arbritary data into zero‐width utf‐8 characters
and decoding it back. These zero‐width characters are not rendered as visible
chracters. These zero‐width characters are then placed inside a message, hiding
the data within the message.

## Compiling

Create the executable 'zwc' in the project directory.
```
make
```

Alternatively,
```
go build -o zwc main/main.go
```

## Installing

The following command should work on Unix-like operating systems.
```
make install
```

## Usage

```
man ./doc/zwc.1
```

or if installed,
```
man zwc
```

## Copyright

Copyright (C) 2023 Ethan Cheng \<ethanrc0528@gmail.com\>  
License GPLv3: GNU GPL version 3 \<http://gnu.org/licenses/gpl.html\>  
This is free software: you are free to change and redistribute it.  
There is NO WARRANTY, to the extent permitted by law.  
