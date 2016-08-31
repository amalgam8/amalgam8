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

package i18n

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/nicksnyder/go-i18n/i18n"
)

// Error JSON
type Error struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

// RestError writes a basic error response with a translated error message and an untranslated error ID
// TODO: request ID?
func RestError(w rest.ResponseWriter, r *rest.Request, code int, id string, args ...interface{}) {
	locale := r.Header.Get("Accept-language")
	T, err := i18n.Tfunc(locale, "en-US")
	if err != nil {
		logrus.WithError(err).WithField(
			"accept_language_header", locale,
		).Error("Could not get translation function")
	}

	translated := T(id, args...)

	errorResp := Error{
		Error:       id,
		Description: translated,
	}

	w.WriteHeader(code)
	w.WriteJson(&errorResp)
}

// LoadLocales loads translation files
func LoadLocales(path string) error {
	logrus.Info("Loading locales")

	filenames, err := filepath.Glob(path + "/*.json")
	if err != nil {
		return err
	}

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

	return nil
}
