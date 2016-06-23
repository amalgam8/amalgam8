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

package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/nicksnyder/go-i18n/i18n"
)

func getQueryIDs(key string, req *rest.Request) []string {
	queries := req.URL.Query()
	values, ok := queries[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	return values
}

func handleDBError(w rest.ResponseWriter, req *rest.Request, err error) {
	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusNotFound {
				RestError(w, req, http.StatusNotFound, "no matching id")
			}
			RestError(w, req, http.StatusServiceUnavailable, "database_error")
		}
		RestError(w, req, http.StatusServiceUnavailable, "failed_to_read_info")
	}
}

func reportMetric(reporter metrics.Reporter, f func(rest.ResponseWriter, *rest.Request) error, name string) rest.HandlerFunc {
	return func(w rest.ResponseWriter, req *rest.Request) {
		startTime := time.Now()
		err := f(w, req)
		endTime := time.Since(startTime)
		if err != nil {
			// Report failure
			reporter.Failure(name, endTime, err)
			return
		}
		// Report success
		reporter.Success(name, endTime)
	}
}

type RestError2 struct {
	ID    string
	Index int
	Error string
	Args  []interface{}
}

func WriteRestErrors(w rest.ResponseWriter, r *rest.Request, restErrors []RestError2, code int) {
	locale := r.Header.Get("Accept-language")
	T, err := i18n.Tfunc(locale, "en-US")
	if err != nil {
		logrus.WithError(err).WithField(
			"accept_language_header", locale,
		).Error("Could not get translation function")
	}

	if len(restErrors) == 0 {
		w.WriteHeader(code)
	} else {
		errorResp := ErrorList{
			Errors: make([]Error, 0, len(restErrors)),
		}
		for _, restError := range restErrors {
			translated := T(restError.Error, restError.Args...)

			errorResp.Errors = append(
				errorResp.Errors,
				Error{
					Index:       restError.Index,
					ID:          restError.ID,
					Error:       restError.Error,
					Description: translated,
				},
			)
		}

		w.WriteHeader(code)
		w.WriteJson(&errorResp)
	}

	return
}

type Error struct {
	ID          string `json:"id,omitempty"`
	Index       int    `json:"index,omitempty"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

type ErrorList struct {
	Errors []Error `json:"errors"`
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
