# Steganography Tool

A CLI Steganography tool that supports input images of format PNG, JPG, BMP and encodes files as PNG with data hidden in color levels. Optional data encryption using `AES-256-GCM` and `argon2id` Key generator.
## Files
* Steg.go   - contains the main function and other primary source code
* REPORT.pdf     - Report about the implementation
* go.mod         - contains dependencies to be downloaded from the internet
* go.sum         - checksum for dependencies
* Data/Data.go   - Package for processing the data to be hidden
* Image/Image.go - Package for processing the data to be hidden
* greetings.gif  - Misc

## Installation

* Dependencies

  * go - https://go.dev/doc/install or `$ sudo snap install go` for Ubuntu users
 
  * chafa -  `$ sudo apt install chafa` (optional, not needed when using --safe)

  * xterm - `$ sudo apt install xterm` (optional, not needed when using --safe)

### Compile and Run
 * Navigate to project Directory
 * `go mod download`
 * `go run Steg.go <arguments>`

 ### Build into Linux executable 
 * Navigate to project Directory
 * `go mod download`
 * `go build Steg.go`

    
## Usage
`./Steg <-e/-d/-c> </path/to/target> </path/to/data> <passphrase> <--safe>`

* -e -- Embedd datafile into target image
* -d -- Extract data from target image and save as file at dir "/path/to/data" 
* -c -- Check hiding capacity of the targetfile

* passphrase -- Optional if data needs to be encrypted
* --safe -- Optional for better stability, use if normal mode is bugged


