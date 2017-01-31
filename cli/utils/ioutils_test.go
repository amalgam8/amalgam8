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

package utils_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/amalgam8/amalgam8/cli/common"
	. "github.com/amalgam8/amalgam8/cli/utils"
	"github.com/amalgam8/amalgam8/pkg/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ioutils", func() {
	fmt.Println("")

	JSONFilePath := "testdata/test_file.json"
	YAMLFilePath := "testdata/test_file.yaml"
	TXTFilePath := "testdata/test_file.txt"

	var _ = BeforeSuite(func() {
		// Check JSON file
		json, JSONerr := ioutil.ReadFile(JSONFilePath)
		Expect(JSONerr).NotTo(HaveOccurred())
		Expect(json).NotTo(BeEmpty())

		// Check YAML file
		yaml, YAMLerr := ioutil.ReadFile(YAMLFilePath)
		Expect(YAMLerr).NotTo(HaveOccurred())
		Expect(yaml).NotTo(BeEmpty())

	})

	Describe("Parsing file", func() {

		Context("when file does not exist", func() {
			It("should return an error", func() {
				_, _, err := ReadInputFile("test_file")
				Expect(err).To(HaveOccurred())
				Expect(err).Should(MatchError(common.ErrFileNotFound))
			})
		})

		Context("when file is a directory", func() {
			It("should return an error", func() {
				_, _, err := ReadInputFile("../utils")
				Expect(err).To(HaveOccurred())
				Expect(err).Should(MatchError(common.ErrInvalidFile))
			})
		})

		Context("when the data type is not JSON or YAML", func() {
			It("should return an error", func() {
				_, format, err := ReadInputFile(TXTFilePath)
				Expect(format).To(Equal("TXT"))
				Expect(err).To(HaveOccurred())
				Expect(err).Should(MatchError(common.ErrUnsoportedFormat))
			})
		})

		Context("when the data is JSON", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(JSONFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(JSON))
				rules := &api.Rule{}

				// Unmarshall JSON Reader
				err = UnmarshallReader(reader, format, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("json_id"))
				Expect(rules.Destination).To(Equal("json_destination"))

				// Marshall reader as JSON
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, format)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("\"id\": \"json_id\""))
			})

		})

		Context("when the data is YAML", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(YAMLFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(YAML))
				rules := &api.Rule{}

				// Unmarshall YAML reader
				err = UnmarshallReader(reader, format, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				// Marshall reader as YAML
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, format)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("id: yaml_id"))
			})

		})

		Context("when converting from YAML to JSON", func() {

			It("should not error", func() {
				reader, format, err := ReadInputFile(YAMLFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(format).To(Equal(YAML))
				rules := &api.Rule{}

				// Convert YAML to JSON
				reader, err = YAMLToJSON(reader, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				// Unmarshall reader as JSON
				err = UnmarshallReader(reader, JSON, rules)
				Expect(err).NotTo(HaveOccurred())
				Expect(rules.ID).To(Equal("yaml_id"))
				Expect(rules.Destination).To(Equal("yaml_destination"))

				//Prettyfy as JSON
				var buf bytes.Buffer
				err = MarshallReader(&buf, rules, JSON)
				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).Should(ContainSubstring("\"id\": \"yaml_id\""))
			})

		})

		Context("when validating JSON rules", func() {

			It("should fix the rules structure", func() {
				reader := bytes.NewBufferString("[]")
				fixed, err := ValidateRulesFormat(reader)
				Expect(err).NotTo(HaveOccurred())

				buf := new(bytes.Buffer)
				buf.ReadFrom(fixed)
				Expect(buf.String()).Should(ContainSubstring("{ \"rules\": []}"))
			})

			It("should not modify the rules structure", func() {
				reader := bytes.NewBufferString("{[]}")
				fixed, err := ValidateRulesFormat(reader)
				Expect(err).NotTo(HaveOccurred())

				buf := new(bytes.Buffer)
				buf.ReadFrom(fixed)
				Expect(buf.String()).Should(ContainSubstring("{[]}"))
			})

		})

		Context("when validating YAML rules", func() {

			It("should fix the rules structure", func() {
				reader := bytes.NewBufferString("-")
				fixed, err := ValidateRulesFormat(reader)
				Expect(err).NotTo(HaveOccurred())

				buf := new(bytes.Buffer)
				buf.ReadFrom(fixed)
				Expect(buf.String()).Should(ContainSubstring("rules:"))
			})

			It("should not modify the rules structure", func() {
				reader := bytes.NewBufferString("rules:")
				fixed, err := ValidateRulesFormat(reader)
				Expect(err).NotTo(HaveOccurred())

				buf := new(bytes.Buffer)
				buf.ReadFrom(fixed)
				Expect(buf.String()).Should(ContainSubstring("rules:"))
			})

		})

	})

})
