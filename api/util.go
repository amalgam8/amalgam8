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

	errorResp := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Error:   id,
		Message: translated,
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
