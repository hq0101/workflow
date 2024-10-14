package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Exec struct {
	WaitFile         string
	PostFile         string
	Outputs          []string
	Command          string
	Args             []string
	TerminationPath  string
	EncodeScriptPath string
}

func (e *Exec) DecodeScript() error {
	scriptFile, err := os.ReadFile(e.EncodeScriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script file %s: %v", e.EncodeScriptPath, err)
	}
	d, err := base64.StdEncoding.DecodeString(string(scriptFile))
	f, err := os.OpenFile(e.EncodeScriptPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open script file %s: %v", e.EncodeScriptPath, err)
	}
	defer f.Close()
	if _, err := f.Write(d); err != nil {
		return fmt.Errorf("failed to write script file %s: %v", e.EncodeScriptPath, err)
	}
	return nil
}

func (e *Exec) Wait() error {
	if e.WaitFile != "" {
		return fmt.Errorf("wait file not implemented")
	}
	for {
		_, err := os.Stat(e.WaitFile)
		if err == nil {
			return nil
		}
		if os.IsNotExist(err) {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return err
	}
}

func (e *Exec) CreatePostFile() error {
	if e.PostFile == "" {
		return fmt.Errorf("post file not implemented")
	}
	if err := os.MkdirAll(filepath.Dir(e.PostFile), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create post file directory %s: %v", e.PostFile, err)
	}
	f, err := os.OpenFile(e.PostFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to create post file %s: %v", e.PostFile, err)
	}
	defer f.Close()
	return nil
}

type output struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e *Exec) readResultsFromDisk() error {
	outputs := []output{}
	for _, result := range e.Outputs {
		if result == "" {
			continue
		}
		file, err := os.ReadFile(filepath.Join("/tmp/results", result))
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		outputs = append(outputs, output{
			Name:  result,
			Value: string(file),
		})
	}

	if err := e.writeTerminationMessage(e.TerminationPath, outputs); err != nil {
		return err
	}

	return nil
}

func (e *Exec) writeTerminationMessage(path string, outputs []output) error {
	fileContents, err := os.ReadFile(path)
	if err == nil {
		var existingEntries []output
		if err := json.Unmarshal(fileContents, &existingEntries); err == nil {
			// append new entries to existing entries
			outputs = append(existingEntries, outputs...)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	jsonOutput, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	if len(jsonOutput) > 1024*4 {
		return fmt.Errorf("termination message too large")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(jsonOutput); err != nil {
		return err
	}
	return f.Sync()
}

func (e *Exec) Run() error {
	cmd := exec.Command(e.Command, e.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if len(e.Outputs) >= 1 && e.Outputs[0] != "" {
		if err := e.readResultsFromDisk(); err != nil {
			return fmt.Errorf("failed to read results from disk: %v", err)
		}
	}

	return nil
}

func main() {
	var encodeScriptPath string
	var waitFile string
	var postFile string
	var command string
	var params string
	var terminationPath string
	var outputs string

	cmd := &cobra.Command{
		Use:   "",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			e := Exec{
				WaitFile:         waitFile,
				PostFile:         postFile,
				Outputs:          strings.Split(outputs, ","),
				Command:          command,
				Args:             strings.Split(params, ","),
				TerminationPath:  terminationPath,
				EncodeScriptPath: encodeScriptPath,
			}
			if err := e.DecodeScript(); err != nil {
				log.Fatalln(err)
			}
			if err := e.Wait(); err != nil {
				log.Fatalln(err)
			}
			if err := e.Run(); err != nil {
				log.Fatalln(err)
			}
			if err := e.CreatePostFile(); err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVarP(&encodeScriptPath, "encode_script", "", "", "")
	cmd.Flags().StringVarP(&waitFile, "wait_file", "", "", "")
	cmd.Flags().StringVarP(&postFile, "post_file", "", "", "")
	cmd.Flags().StringVarP(&command, "command", "", "/bin/sh", "")
	cmd.Flags().StringVarP(&outputs, "outputs", "", "", "")
	cmd.Flags().StringVarP(&terminationPath, "termination_path", "", "/tmp/termination-log", "")
	cmd.Flags().StringVarP(&params, "params", "", "", "")
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
