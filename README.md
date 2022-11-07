# Delta
FUD Linux Remote Access Trojan

## Capabilities : 
- Execute shell commands
- Capture screenshot of victim machine
- Download files from victim machine
- Upload files to victim machine

## Set-Up :
1. Clone this repository
2. Navigate to the project directory
3. Feel free to modify C2 Address in the delta.go file to point to your C2
4. Run the command ```go build delta.go``` to create the executable(ELF) file

## Usage :
### NOTE : 
A C2 Server for Delta is currently in development and will be released soon ! \
In the meantime, you can use tools like netcat though you won't have the convenience of enjoying all the functionality present in the RAT

