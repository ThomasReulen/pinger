package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	Time               string
	RoundTripMin       string
	RoundTripAverage   string
	RoundTripMax       string
	RoundTripDeviation string
	Warning            string
}

// PingReply contains an individual ping reply line.
type JsonReplies struct {
	Size           string
	FromAddress    string
	SequenceNumber string
	TTL            string
	Time           string
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
func Ping(ipV4Address string, count int) (*parser.PingOutput, error) {
	var (
		output, errorOutput bytes.Buffer
		exitCode            int
	)
	pingArgs := []string{"-c", fmt.Sprintf("%d", count), ipV4Address}
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

	err := os.Mkdir("data", 0755)
	if err != nil {
		log.Println(err)
	}

	ip := os.Getenv("IP")
	fmt.Printf("%s\n", ip)

	inst := fmt.Sprint(time.Now().Unix())
	fmt.Printf("%s\n", inst)

	datafolder := os.Getenv("DATA_FOLDER")
	if len(datafolder) == 0 {
		datafolder = "data"
	}
	err = os.Mkdir(datafolder, 0755)
	if err != nil {
		log.Print(err)
	}

	chunksize := os.Getenv("CHUNKSIZE")
	if len(chunksize) == 0 {
		chunksize = "3"
	}
	cs, _ := strconv.Atoi(chunksize)

	iterations := os.Getenv("ITERATIONS")
	if len(iterations) == 0 {
		iterations = "1"
	}
	its, _ := strconv.Atoi(iterations)

	for i := 0; i < its; i++ {

		p, err := Ping(ip, cs)
		if err != nil {
			fmt.Printf("%v", err)
		}

		pingRes := JsonPingResult{
			Host:              p.Host,
			ResolvedIPAddress: p.ResolvedIPAddress,
			PayloadSize:       UintToStr(p.PayloadSize),
			PayloadActualSize: UintToStr(p.PayloadActualSize),
		}

		ps := JsonPingStatistics{
			IPAddress:          p.Stats.IPAddress,
			PacketsTransmitted: fmt.Sprint(p.Stats.PacketsTransmitted),
			PacketsReceived:    fmt.Sprint(p.Stats.PacketsReceived),
			Errors:             fmt.Sprint(p.Stats.Errors),
			PacketLossPercent:  fmt.Sprint(p.Stats.PacketLossPercent),
			Time:               p.Stats.Time.String(),
			RoundTripMin:       p.Stats.RoundTripMin.String(),
			RoundTripAverage:   p.Stats.RoundTripAverage.String(),
			RoundTripMax:       p.Stats.RoundTripMax.String(),
			RoundTripDeviation: p.Stats.RoundTripDeviation.String(),
			Warning:            p.Stats.Warning,
		}
		pingRes.Stats = ps

		for _, pr := range p.Replies {
			singleReply := JsonReplies{
				Size:           fmt.Sprint(pr.Size),
				FromAddress:    pr.FromAddress,
				SequenceNumber: fmt.Sprint(pr.SequenceNumber),
				TTL:            fmt.Sprint(pr.TTL),
				Time:           pr.Time.String(),
				Error:          pr.Error,
				Duplicate:      pr.Duplicate,
			}
			pingRes.Replies = append(pingRes.Replies, singleReply)
		}
		runstamp := fmt.Sprint(time.Now().Unix())
		jsonPackage, _ := json.Marshal(pingRes)
		os.WriteFile(datafolder+"/"+runstamp+".json", jsonPackage, 0666)
	}
}
