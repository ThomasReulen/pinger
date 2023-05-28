package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"parser"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type JsonPingStatistics struct {
	IPAddress          string
	PacketsTransmitted string
	PacketsReceived    string
	Errors             string
	PacketLossPercent  string
	Time               time.Duration
	RoundTripMin       time.Duration
	RoundTripAverage   time.Duration
	RoundTripMax       time.Duration
	RoundTripDeviation time.Duration
	Warning            string
}

// PingReply contains an individual ping reply line.
type JsonReplies struct {
	Size           string
	FromAddress    string
	SequenceNumber string
	TTL            string
	Time           time.Duration
	Error          string
	Duplicate      bool
}

type JsonPingResult struct {
	Host              string
	ResolvedIPAddress string
	PayloadSize       string
	PayloadActualSize string
	Replies           []JsonReplies
	Stats             JsonPingStatistics
}

func UintToStr(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// Ping will ping the specified IPv4 address wit the provided timeout, interval and size settings .
func Ping(ipV4Address string, interval, timeout time.Duration, size uint) (*parser.PingOutput, error) {
	var (
		output, errorOutput bytes.Buffer
		exitCode            int
	)

	pingArgs := []string{"-n", "-s", fmt.Sprintf("%d", size), "-w", fmt.Sprintf("%d", int(timeout.Seconds())), "-i", fmt.Sprintf("%d", int(interval.Seconds())), ipV4Address}
	cmd := exec.Command("ping", pingArgs...)
	cmd.Stdout = &output
	cmd.Stderr = &errorOutput

	err := cmd.Run()
	if err == nil {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	} else {
		exitCode, err = parseExitCode(err)
		if err != nil {
			return nil, err
		}
	}

	// try to parse output also in case of failure
	po, err := parser.Parse(output.String())
	if err == nil {
		return po, nil
	}

	// in case of error, use also the execution context errors (if any)
	return nil, fmt.Errorf("command: ping %s\nexit code: %d\nparse error: %v\nstdout:\n%s\nstderr:\n%s", strings.Join(pingArgs, " "), exitCode, err, output.String(), errorOutput.String())
}

func parseExitCode(err error) (int, error) {
	// try to get the exit code
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		return ws.ExitStatus(), nil
	}

	// This will happen (in OSX) if `name` is not available in $PATH,
	// in this situation, exit code could not be get, and stderr will be
	// empty string very likely, so we use the default fail code, and format err
	// to string and set to stderr
	return 0, fmt.Errorf("could not get exit code for failed program: %v", err)
}

func main() {

	ip := "8.8.8.8"
	for i := 0; i < 10; i++ {
		p, err := Ping(ip, time.Second, time.Second*3, 128)
		if err != nil {
			fmt.Printf("%v", err)
		}
		jsonPackage, _ := json.Marshal(p)
		fmt.Printf("%s", jsonPackage)
		if i > 3 {
			ip = "192.168.199.253"
		} else {
			ip = "8.8.8.8"
		}
	}
	// pingRes := JsonPingResult{
	// 	Host:              p.Host,
	// 	ResolvedIPAddress: p.ResolvedIPAddress,
	// 	PayloadSize:       UintToStr(p.PayloadSize),
	// 	PayloadActualSize: UintToStr(p.PayloadActualSize),
	// }
	// for _, pr := range p.Replies {
	// 	// fmt.Println(pr.Duplicate, pr.Error, pr.FromAddress, pr.SequenceNumber)
	// 	// fmt.Println(json.Marshal(pr))
	// }
	// fmt.Printf("%v", p.Stats.RoundTripMax)
	// fmt.Printf("%s", pingRes)

}
