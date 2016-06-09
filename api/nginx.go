package api

import (
	"bytes"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/nginx"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/nicksnyder/go-i18n/i18n"
)

// NGINXConfig options
type NGINXConfig struct {
	Statsd    statsd.Statter
	Generator nginx.Generator
	Checker   checker.Checker
}

// NGINX handles NGINX API calls
type NGINX struct {
	statsd    statsd.Statter
	generator nginx.Generator
	checker   checker.Checker
}

// NewNGINX creates struct
func NewNGINX(nc NGINXConfig) *NGINX {
	return &NGINX{
		statsd:    nc.Statsd,
		generator: nc.Generator,
		checker:   nc.Checker,
	}
}

// Routes for NGINX API calls
func (n *NGINX) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/tenants/#id/nginx", ReportMetric(n.statsd, n.GetNGINX, "tenants_nginx")),
	}
}

// ReportMetric TODO
func ReportMetric(client statsd.Statter, f func(rest.ResponseWriter, *rest.Request) error, name string) rest.HandlerFunc {
	return func(w rest.ResponseWriter, req *rest.Request) {
		startTime := time.Now()
		err := f(w, req)
		endTime := time.Since(startTime)
		if err != nil {
			logrus.WithError(err).Error("API failed")
			// Report failure
			client.Inc(name+"CountFailure", 1, 1.0)
			client.TimingDuration(name+"ResponseTimeFailure", endTime, 1.0)
			return
		}
		// Report success
		client.Inc(name+"CountSuccess", 1, 1.0)
		client.TimingDuration(name+"ResponseTimeSuccess", endTime, 1.0)
	}
}

// GetNGINX returns the NGINX configuration for a given tenant
func (n *NGINX) GetNGINX(w rest.ResponseWriter, req *rest.Request) error {
	var err error

	id := req.PathParam("id")
	queries := req.URL.Query()
	var lastUpdate *time.Time
	if queries.Get("version") != "" {
		update, err := time.Parse(time.RFC3339, queries.Get("version"))
		if err == nil {
			lastUpdate = &update
		}
	}

	catalog, err := n.checker.Get(id)
	if err != nil {
		if ce, ok := err.(*database.DBError); ok {
			if ce.StatusCode == http.StatusNotFound {
				RestError(w, req, http.StatusNotFound, "no matching id")
				return err
			}
			RestError(w, req, http.StatusServiceUnavailable, "rules_database_error")
			return err
		}
		RestError(w, req, http.StatusServiceUnavailable, "get_rules_failed")
		return err
	}

	// if version query is newer than latest rules change, return 204
	if lastUpdate != nil && catalog.LastUpdate.Before(*lastUpdate) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Generate config
	buf := bytes.NewBuffer([]byte{})
	if err = n.generator.Generate(buf, id); err != nil {
		RestError(w, req, http.StatusInternalServerError, "error_nginx_generator_failed")
		return err
	}

	// Write response as text
	httpWriter, ok := w.(http.ResponseWriter)
	if !ok {
		RestError(w, req, http.StatusInternalServerError, "error_internal")
		return errors.New("Could not cast rest.ResponseWriter to http.ResponseWriter")
	}
	httpWriter.Header().Set("Content-type", "text/plain")
	httpWriter.WriteHeader(http.StatusOK)
	httpWriter.Write(buf.Bytes())

	return nil
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
