package main

import (
	"fmt"
	"github.com/peterh/liner"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	kubeConfigCmd = "--kubeconfig /etc/rancher/k3s/k3s.yaml"
	history_fn    = filepath.Join(os.TempDir(), ".liner_example_history")
	keywords      = []string{"kubectl", "create", "update", "delete", "deployment"}
)

const (
	downloadPromptValue = " Kubectl Demo will download k3s (lighweight kubernetes detail in https://k3s.io/ ) \n" +
		"and create systemctl service approximately 40mb file will download." +
		"Do you agree with this  [y/N] ? "
	resetK3sPromtValue = "kubectl demo will remove your k3s instance. Do you want to continue [y/N] ? "
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		initK3s()
	} else {
		switch argsWithoutProg[0] {

		case "uninstall", "remove", "delete":
			if Confirm(resetK3sPromtValue) {
				uninstallK3s()
			}
			initK3s()

		}
	}
}
func uninstallK3s() {

	stopServer()
	commandRun("/usr/local/bin/k3s-uninstall.sh")

}

func initK3s() {

	if !isInstalled() {
		fmt.Println("Starting K3s installation")
		installK3s()
	}
	if startK3sServer() {
		fmt.Println("Server succesfully started")
	} else {
		fmt.Println("Server start error. You could check details in k3s service log ")
		return
	}

	//loadDemoData()

	createTerminal()

	defer stopServer()
}
func createTerminal() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range keywords {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	if f, err := os.Open(history_fn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	for {
		if cmdString, err := line.Prompt("kubectl-demo$ "); err == nil {
			line.AppendHistory(cmdString)
			if strings.Contains(cmdString, "exit") || strings.Contains(cmdString, "quit") {
				return
			}
			err = runCommand(cmdString)
		} else if err == liner.ErrPromptAborted {
			return
		} else {
			log.Print("Error reading line: ", err)
			return
		}
		if f, err := os.Create(history_fn); err != nil {
			log.Print("Error writing history file: ", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}

func runCommand(commandStr string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	if strings.Contains(commandStr, "kubectl") {
		commandStr = commandStr + " " + kubeConfigCmd
	}
	arrCommandStr := strings.Fields(commandStr)

	if len(arrCommandStr) > 1 {
		cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		return cmd.Run()
	} else if len(arrCommandStr) == 1 {
		cmd := exec.Command(arrCommandStr[0])
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}
	return nil
}

func stopServer() {
	fmt.Println("Stopping K3s server")
	err := commandRun("systemctl stop k3s")
	if err != nil {
		fmt.Printf(" Server start failed %v\n", err)
		return
	}
}

func startK3sServer() bool {
	fmt.Println("Starting K3s server")
	err := commandRun("systemctl start k3s ")
	if err != nil {
		fmt.Printf(" Server start failed %v\n", err)
		return false
	}
	return serverHealth()

}
func serverHealth() bool {

	for start := time.Now(); time.Since(start) < time.Second*10; {
		cmd := exec.Command("systemctl", "check", " k3s")
		bytes, _ := cmd.CombinedOutput()

		if strings.Contains(string(bytes), "active") {
			return true
		} else {
			time.Sleep(time.Second)
		}
	}
	return false

}

func installK3s() {
	if Confirm(downloadPromptValue) {
		err := commandRun("cd " + os.TempDir() + " && curl -sfL https://get.k3s.io | sh -")
		if err != nil {
			fmt.Printf("Download k3s failed please try again %v\n", err)
		}
	} else {
		fmt.Println("Kubectl Demo is just useless without k3s. \n " +
			"=============================================== \n  " +
			"====> May the Kubernetes be with you :) <====== \n " +
			"=============================================== \n ")
		os.Exit(0)
	}
}

func isInstalled() bool {
	cmd := exec.Command("systemctl", "status", "k3s")
	bytes, _ := cmd.CombinedOutput()
	if strings.Contains(string(bytes), "could not be found") {
		fmt.Println("installation could not be found")
		return false
	} else if strings.Contains(string(bytes), "active") || strings.Contains(string(bytes), "k3s.io") {
		fmt.Println("installation and service active")
		return true
	} else {
		fmt.Println("installation error" + string(bytes))
		return false
	}
}

func Confirm(promptValue string, args ...interface{}) bool {
	for {
		switch Prompt(promptValue, args...) {
		case "Yes", "yes", "y", "Y":
			return true
		case "No", "no", "n", "N":
			return false
		}
	}
}

func Prompt(prompt string, args ...interface{}) string {
	var s string
	fmt.Printf(prompt+": ", args...)
	fmt.Scanln(&s)
	return s
}

func commandRun(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
