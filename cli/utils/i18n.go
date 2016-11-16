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
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/nicksnyder/go-i18n/i18n"
)

// LoadLocales loads translation files
func LoadLocales(path string) error {
	logrus.Info("Loading locales")

	if path == "" {
		path = "locales"
	}
	filenames, err := filepath.Glob(path + "/*.json")
	if err != nil {
		return err
	}

	// For development use local files intead of the compiled resources.
	// run "go-bindata -pkg=utils -prefix "./cli" -o ./cli/utils/i18n_resources.go ./cli/locales" to compile i18n
	if len(filenames) > 0 {
		for _, filename := range filenames {
			logrus.Debug(filename)
			filename, err = filepath.Abs(filename)
			if err != nil {
				return err
			}

			logrus.Debug(filename)
			if err = i18n.LoadTranslationFile(filename); err != nil {
				return err
			}
		}
	} else {
		// TODO: enable other languages
		filename := filepath.Join("locales", "en-US.json")
		data, err := Asset(filename)
		if err != nil {
			return err
		}

		if err = i18n.ParseTranslationFileBytes(filename, data); err != nil {
			return err
		}
	}

	return nil
}

// Language is a wrapper of go_i18n.Tfunc
func Language(languageSource string, languageSources ...string) i18n.TranslateFunc {
	T, err := i18n.Tfunc(languageSource)
	if err != nil {
		logrus.Error(err)
	}
	return T
}
