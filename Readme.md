# Legacy GoBit
(Obituaries retrieved from Legacy website, with the go programming language.)
This retrieves Obituaries from Legacy for a given base area, sending the results via Gmail.

This will store the links retrieved in a basic text file so that the next time it's run, it 
won't pick up the same obituaries.

## Note
This is a quick project for me to get up to speed on the Go programming language, which (as of August, 2019), I'm still very new.

What I've learned from this:
* Sqlite-related CRUD
* Exception handling
* Slices/Array logic
* Map handling and use-cases (for uniqueness)
* Text file I/O
* Struct / Inner Structs
* JSON parsing

What's not done
* Tests
* Sqlite stuff (was commented out due to non-portability to Linux systems that had Sqlite bindings compiled differently)
* Channel stuff - I really wanted to do concurrent processing, but had to the draw the line somewhere.

## Usage
`$ LegacyGoBit -t <to email list separated by ,> -f <from email> -p <password> -u <Legacy.com base URL>`

