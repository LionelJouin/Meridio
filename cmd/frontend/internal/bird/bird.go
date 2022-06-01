/*
Copyright (c) 2021 Nordix Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bird

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var regexWarn *regexp.Regexp = regexp.MustCompile(`Error|<ERROR>|<BUG>|<FATAL>|<WARNING>`)
var regexInfo *regexp.Regexp = regexp.MustCompile(`<INFO>|BGP session|Connected|Received:|Started|Neighbor|Startup delayed`)

func NewService(commSocket string, configFile string) *Service {
	return &Service{
		communicationSocket: commSocket,
		configFile:          configFile,
	}
}

type Service struct {
	communicationSocket string // filename (with path) to communicate with birdc
	configFile          string // configuration file (with path)
}

// LookupCli -
// Looks up birdc path
func (b *Service) LookupCli() (string, error) {
	path, err := exec.LookPath("birdc")
	if err != nil {
		err = errors.New("Birdc not found, err: " + err.Error())
	}
	return path, err
}

// CliCmd -
// Executes birdc commands
func (b *Service) CliCmd(ctx context.Context, lp string, arg ...string) (string, error) {
	if lp == "" {
		path, err := b.LookupCli()
		if err != nil {
			return path, err
		}
		lp = path
	}

	arg = append([]string{"-s", b.communicationSocket}, arg...)
	cmd := exec.CommandContext(ctx, lp, arg...)
	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}

// CheckCli -
// Checks if birdc is available
func (b *Service) CheckCli(ctx context.Context, lp string) error {
	if lp == "" {
		path, err := b.LookupCli()
		if err != nil {
			return err
		}
		lp = path
	}

	cmd := exec.CommandContext(ctx, lp, "-s", b.communicationSocket, "show", "status")
	stdoutStderr, err := cmd.CombinedOutput()
	stringOut := string(stdoutStderr)
	if err != nil {
		return errors.New(cmd.String() + ": " + err.Error() + ": " + stringOut)
	}
	return nil
}

// Run -
// Starts BIRD process with the config file (blocks)
// Based on monitorLogs settings stderr of the started BIRD process can be monitored,
// in order to append important log snippets to the container's log
func (b *Service) Run(ctx context.Context, monitorLogs bool) error {
	if !monitorLogs {
		if stdoutStderr, err := exec.CommandContext(ctx, "bird", "-d", "-c", b.configFile, "-s", b.communicationSocket).CombinedOutput(); err != nil {
			logrus.Errorf("BIRD Start: err: \"%v\", out: %s", err, stdoutStderr)
			return err
		}
	} else {
		cmd := exec.CommandContext(ctx, "bird", "-d", "-c", b.configFile, "-s", b.communicationSocket)
		// get stderr pipe reader that will be connected with the process' stderr by Start()
		pipe, err := cmd.StderrPipe()
		if err != nil {
			logrus.Errorf("BIRD Start: stderr pipe err: \"%v\"", err)
			return err
		}

		// Note: Probably not needed at all, as due to the use of CommandContext()
		// Start() would kill the process as soon context becomes done. Which should
		// lead to an EOF on stderr anyways.
		go func() {
			// make sure bufio Scan() can be breaked out from when context is done
			w, ok := cmd.Stderr.(*os.File)
			if !ok {
				// not considered a deal-breaker at the moment; see above note
				logrus.Debugf("BIRD Start: cmd.Stderr not *os.File")
				return
			}
			// when context is done, close File thus signalling EOF to bufio Scan()
			defer w.Close()
			<-ctx.Done()
			logrus.Infof("BIRD Start: context closed, terminate log monitoring...")
		}()

		// start the process (BIRD)
		if err := cmd.Start(); err != nil {
			logrus.Errorf("BIRD Start: start err: \"%v\"", err)
			return err
		}
		if err := b.monitorOutput(pipe); err != nil {
			logrus.Errorf("BIRD Start: scanner err: \"%v\"", err)
			return err
		}
		// wait until process concludes
		// (should only get here after stderr got closed or scanner returned error)
		if err := cmd.Wait(); err != nil {
			logrus.Errorf("BIRD Start: err: \"%v\"", err)
			return err
		}
	}
	return nil
}

// ShutDown -
// Shuts BIRD down via birdc
func (b *Service) ShutDown(ctx context.Context, lp string) error {
	out, err := b.CliCmd(ctx, lp, "down")
	if err != nil {
		err = fmt.Errorf("%v - %v", err, out)
	}
	return err
}

// monitorOutput -
// Keeps reading the output of a BIRD process and adds important
// log snippets to the containers log for debugging purposes
func (b *Service) monitorOutput(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if ok := regexWarn.MatchString(scanner.Text()); ok {
			logrus.Warnf("[bird] %v", scanner.Text())
		} else if ok := regexInfo.MatchString(scanner.Text()); ok {
			logrus.Infof("[bird] %v", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Configure -
// Orders BIRD via birdc to (try and) apply the config
func (b *Service) Configure(ctx context.Context, lp string) (string, error) {
	arg := `"` + b.configFile + `"`
	stringOut, err := b.CliCmd(ctx, lp, "configure", arg)

	if err != nil {
		return stringOut, err
	} else if !strings.Contains(stringOut, ReconfigInProgress) && !strings.Contains(stringOut, Reconfigured) {
		return stringOut, errors.New("reconfiguration failed")
	} else {
		return stringOut, nil
	}
}

// Verify -
// Verifies content of config file via birdc
func (b *Service) Verify(ctx context.Context, lp string) (string, error) {
	arg := `"` + b.configFile + `"`
	stringOut, err := b.CliCmd(ctx, lp, "configure", "check", arg)
	if err != nil {
		return stringOut, err
	} else if !strings.Contains(stringOut, ConfigurationOk) {
		return stringOut, errors.New("verification failed")
	} else {
		return stringOut, nil
	}
}

// ShowProtocolSessions -
// Retrieves detailed routing protocol information via birdc
func (b *Service) ShowProtocolSessions(ctx context.Context, lp, pattern string) (string, error) {
	args := []string{
		"show",
		"protocols",
		"all",
	}
	if pattern != "" {
		args = append(args, `"`+pattern+`"`)
	}
	return b.CliCmd(ctx, lp, args...)
}

// ShowProtocolSessions -
// Retrieves information on the available BFD sessions (for the given BFD protocol name if any)
func (b *Service) ShowBfdSessions(ctx context.Context, lp, name string) (string, error) {
	args := []string{
		"show",
		"bfd",
		"session",
	}
	if name != "" {
		args = append(args, `'`+name+`'`)
	}
	return b.CliCmd(ctx, lp, args...)
}
