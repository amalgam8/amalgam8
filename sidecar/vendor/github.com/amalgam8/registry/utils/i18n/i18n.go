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

// Package i18n provides data structures and functions to support externalizing user visible messages from the code,
// thus allowing them to be translated in accordance with locale requirements
package i18n

import (
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/nicksnyder/go-i18n/i18n"
)

// LoadLocales for globalization support
func LoadLocales(path string) error {
	log.WithFields(log.Fields{
		"path": path,
	}).Info("Loading language files from directory")

	filenames, err := filepath.Glob(path + "/*.json")
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"path": filename,
		}).Debug("Loading locale")
		if err = i18n.LoadTranslationFile(filename); err != nil {
			return err
		}
	}

	return nil
}

// TranslateFunc determines what translation to load. For the common case, calling i18n.Error() should suffice
func TranslateFunc(req *rest.Request) i18n.TranslateFunc {
	const (
		acceptLanguage = "Accept-Language"
		requestID      = "SD-Request-ID" // copied to avoid cyclic dependency between middleware and i18n
	)

	// Using golang.org/x/text/language may be a better option for matching languages (e.g, support for language
	// weights, see https://godoc.org/golang.org/x/text/language#example-ParseAcceptLanguage), but we rely on the
	// internal implementation in github.com/nicksnyder/go-i18n/i18n to handle all that for us...
	reqID := req.Header.Get(requestID)
	locale := req.Header.Get(acceptLanguage)
	T, err := i18n.Tfunc(locale, "en-US")
	if err != nil {
		log.WithFields(log.Fields{
			"error":                  err,
			"request_id":             reqID,
			"accept_language_header": locale,
		}).Error("Could not get translation function")
	}
	return T
}

// Error produces an error response in JSON with the following structure, '{"Error":"error message"}',
// where the error message is the translation corresponding to 'id', parameterized by 'args' (if present)
func Error(r *rest.Request, w rest.ResponseWriter, code int, id string, args ...interface{}) {
	T := TranslateFunc(r)
	translated := T(id, args...)
	rest.Error(w, translated, code)
}

// SupressTestingErrorMessages loads a minimal en-US locale for testing purposes only. Should be called in init()
func SupressTestingErrorMessages() {
	_ = i18n.ParseTranslationFileBytes("en-US.json", []byte(`[{"id":"test", "translation":"message"}]`))
}
