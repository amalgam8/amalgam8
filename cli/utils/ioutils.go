// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/amalgam8/amalgam8/cli/common"
	"gopkg.in/yaml.v2"
)

const (
	// JSON .
	JSON = "JSON"
	// YAML .
	YAML = "YAML"
)

// ReadInputFile reads the content of a given file and retuns a reader and the extension of the file.
func ReadInputFile(path string) (io.Reader, string, error) {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", common.ErrFileNotFound
	}
	// Check if it's a folder
	if info.IsDir() {
		return nil, "", common.ErrInvalidFile
	}

	// Get the file extension
	ext := strings.ToUpper(filepath.Ext(path)[1:])
	if ext != JSON && ext != YAML {
		return nil, ext, common.ErrUnsoportedFormat
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, ext, err
	}

	return bytes.NewReader(data), ext, nil
}

//YAMLToJSON converts the YAML data of a given reader to JSON, removes all insignificant characters and returns a new reader.
func YAMLToJSON(reader io.Reader, dst interface{}) (io.Reader, error) {
	err := UnmarshalReader(reader, YAML, dst)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = MarshallReader(&buf, dst, JSON)
	if err != nil {
		return nil, err
	}

	var compact bytes.Buffer
	err = json.Compact(&compact, buf.Bytes())
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(compact.Bytes()), nil
}

// UnmarshalReader parses the reader data and stores the result in the value pointed by dest.
func UnmarshalReader(data io.Reader, format string, dest interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(data)
	format = strings.ToUpper(format)
	if format != JSON && format != YAML {
		return common.ErrUnsoportedFormat
	}
	switch format {

	case JSON:
		err := json.Unmarshal(buf.Bytes(), dest)
		if err != nil {
			return err
		}

	case YAML:
		err := yaml.Unmarshal(buf.Bytes(), dest)
		if err != nil {
			return err
		}

	default:
		return common.ErrInvalidFormat
	}

	return nil
}

// MarshallReader returns the JSON or YAML encoding of data.
func MarshallReader(writer io.Writer, data interface{}, format string) error {
	format = strings.ToUpper(format)
	if format != JSON && format != YAML {
		return common.ErrUnsoportedFormat
	}
	var pretty []byte
	var err error
	switch format {
	case JSON:
		pretty, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}

	case YAML:
		pretty, err = yaml.Marshal(data)
		if err != nil {
			return err
		}

	default:
		return common.ErrInvalidFormat
	}

	// escape "<", ">" and "&"
	pretty = bytes.Replace(pretty, []byte("\\u003c"), []byte("<"), -1)
	pretty = bytes.Replace(pretty, []byte("\\u003e"), []byte(">"), -1)
	pretty = bytes.Replace(pretty, []byte("\\u0026"), []byte("&"), -1)

	fmt.Fprintf(writer, "\n%+v\n\n", string(pretty))
	return nil
}

// ScannerLines returns a reader with the rules provided in the console.
func ScannerLines(writer io.Writer, description string, instructions bool) (io.Reader, string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	dataBuf := bytes.Buffer{}
	var format string
	breakers := map[string]string{
		".json": JSON,
		".yaml": YAML,
	}

	if !instructions {
		fmt.Fprintf(writer, "\n%s\n%s:\n\n", "#########################################################################\nPress (CTRL+D) on Unix-like systems or (Ctrl+Z) in Windows when finished.\nYou can also write .json or .yaml in a new line to finish.\n#########################################################################\n", description)
	}

	for scanner.Scan() {
		for k := range breakers {
			if strings.ToLower(scanner.Text()) == k {
				format = breakers[k]
				break
			}
		}
		if format != "" {
			break
		}
		fmt.Fprintln(&dataBuf, scanner.Text())
	}

	err := scanner.Err()
	if err != nil {
		return nil, format, err
	}

	// If buffer is empty
	if dataBuf.Len() == 0 {
		return nil, format, errors.New("No Rules")
	}

	fixedReader, err := ValidateRulesFormat(&dataBuf)
	if err != nil {
		return nil, format, err
	}

	dataBuf.Reset()
	dataBuf.ReadFrom(fixedReader)

	// When format is empty, probably EOF was used.
	// findFormat() will try to guess the format of the rules.
	if format == "" {
		format, err = findFormat(dataBuf)
		if err != nil {
			return nil, format, err
		}
	}

	return &dataBuf, format, nil
}

// Confirmation .
func Confirmation(writer io.Writer, description string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprintf(writer, "%s [y/N]: ", description)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true, nil
		}
		return false, nil
	}
}

func findFormat(buf bytes.Buffer) (string, error) {
	var data map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &data) == nil {
		return JSON, nil
	} else if yaml.Unmarshal(buf.Bytes(), &data) == nil {
		return YAML, nil
	}
	return "", common.ErrInvalidFormat
}

// ValidateRulesFormat .
func ValidateRulesFormat(reader io.Reader) (io.Reader, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	format := YAML
	// if the buffer begin with "[" or "{" assume that the content should be in JSON format
	if strings.HasPrefix(buf.String(), "[") || strings.HasPrefix(buf.String(), "{") {
		format = JSON
	}

	fixedBuf := &bytes.Buffer{}
	switch format {
	case YAML:
		if !strings.HasPrefix(buf.String(), "rules:") {
			fixedBuf.WriteString(fmt.Sprint("rules:\n", buf.String()))
		}
	case JSON:
		if strings.HasPrefix(buf.String(), "[") {
			fixedBuf.WriteString(fmt.Sprint("{ \"rules\": ", buf.String(), "}"))
		}
	default:
		return nil, common.ErrInvalidFormat
	}

	if len(fixedBuf.String()) != 0 {
		return fixedBuf, nil
	}

	return buf, nil
}
