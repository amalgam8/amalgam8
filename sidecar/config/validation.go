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

package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/miekg/dns"
)

// Validate runs validation checks
func Validate(validators []ValidatorFunc) error {
	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}
	return nil
}

// ValidatorFunc performs a validation check
type ValidatorFunc func() error

// IsNotEmpty ensures the value is not empty
func IsNotEmpty(name, value string) ValidatorFunc {
	return func() error {
		if value == "" {
			return errors.New(name + " is empty")
		}
		return nil
	}
}

// IsValidURL ensures that the URL is valid and has a protocol specified (http, https, etc...)
func IsValidURL(name, value string) ValidatorFunc {
	return func() error {
		if value == "" {
			return errors.New(name + " is empty")
		}

		u, err := url.Parse(value)
		if err != nil {
			return errors.New(name + " is not a valid URL: " + value)
		}

		if u.Scheme == "" || u.Host == "" {
			return errors.New(name + " is not a valid URL: " + value)
		}

		return nil
	}
}

// IsEmptyOrValidURL ensures that the URL is either empty, or is a valid URL
func IsEmptyOrValidURL(name, value string) ValidatorFunc {
	return func() error {
		if value == "" {
			return nil
		}
		return IsValidURL(name, value)()
	}
}

// IsInRange ensures the integer is within the specified inclusive range
func IsInRange(name string, value, min, max int) ValidatorFunc {
	return func() error {
		if value < min || value > max {
			return fmt.Errorf("%v is not in range [%v, %v]", name, min, max)
		}
		return nil
	}
}

// IsInSet ensures that the value is a member of the set
func IsInSet(name, value string, set []string) ValidatorFunc {
	return func() error {
		for _, item := range set {
			if value == item {
				return nil
			}
		}

		return fmt.Errorf("%v is not a member of the set %v", name, set)
	}
}

// IsInRangeDuration ensures the time value is between the given min and max times
func IsInRangeDuration(name string, value, min, max time.Duration) ValidatorFunc {
	return func() error {
		if value.Seconds() < min.Seconds() || value.Seconds() > max.Seconds() {
			return fmt.Errorf("%v must be more than %v and less than %v", name, min, max)
		}
		return nil
	}
}

// IsValidDomain ensures the domain name is valid domain and that it contains only one domain.
func IsValidDomain(name, value string) ValidatorFunc {
	return func() error {
		numberOfDomains, isValidDomain := dns.IsDomainName(value)
		if !isValidDomain {
			return errors.New(name + " not a valid domain")
		}
		if numberOfDomains != 1 {
			return errors.New(name + " needs to contain 1 domain")
		}
		return nil
	}
}
