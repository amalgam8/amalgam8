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

package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/api/env"
)

var globalTraceNameGenerator = NewNameGenerator(defaultDictionary)

// Trace provides mechanism to add headers to each response returned by the server
// the implementation leaves to the user to specify the header name/label and format the value as any string
// may it be a comma separated multi-value string.
type Trace struct {
	name     string
	headers  http.Header
	curIndex uint64
}

// NewTrace creates and initialize a trace middleware
func NewTrace() *Trace {
	return &Trace{
		name:    globalTraceNameGenerator.Generate(10),
		headers: make(http.Header),
	}
}

// AddHeader adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (mw *Trace) AddHeader(key, value string) {
	mw.headers.Add(key, value)
}

// MiddlewareFunc returns a go-json-rest HTTP Handler function, wrapping calls to the provided HandlerFunc
func (mw *Trace) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(writer rest.ResponseWriter, request *rest.Request) { mw.handler(writer, request, handler) }
}

func (mw *Trace) handler(w rest.ResponseWriter, r *rest.Request, h rest.HandlerFunc) {
	index := mw.nextIndex()
	traceID := fmt.Sprintf("%s_%s_%d", mw.name, strconv.FormatInt(time.Now().UnixNano(), 10), index)

	r.Env[env.RequestID] = traceID

	for headerName, headerValue := range mw.headers {
		for _, value := range headerValue {
			w.Header().Set(headerName, value)
		}
	}

	w.Header().Set(env.RequestID, traceID)

	// Important, Need to call next handler last, since following middleware activate Write() to the connection,
	// requiring us to add the header at the pre stage of middleware chain
	h(w, r)

}

func (mw *Trace) nextIndex() uint64 {
	return atomic.AddUint64(&mw.curIndex, 1)
}

//
// TODO: We can extract the following to utils package
//

// NameGenerator generates unique sequence of strings from the provided dictionary
type NameGenerator interface {
	// Generate a sequence of n characters
	Generate(n int) string
}

type nameGenerator struct {
	dictionary string
	// Number of bits required to represent the index in the dictionary
	numBits uint
	// All 1-numBits, necessary to mask out numBits
	mask int64
	// Number of dictionary indices fitting in 63 bits
	maxDictionaryIndices uint

	randGen rand.Source

	sync.RWMutex
}

const defaultDictionary = "abcABCdefDEFghiGHIjklJKLmnopMNOPqrstuvwxyzQRSTUVWXYZ"

// NewNameGenerator creates a new generator of random string sequences from the requested dictionary
func NewNameGenerator(dictionary string) NameGenerator {
	if dictionary == "" {
		dictionary = defaultDictionary
	}

	sizeOfDictionary := len(dictionary)
	// Compute Number of bits required to represent the dictionary
	numBits := uint(0)
	for pos, shifts := int(1), uint(0); sizeOfDictionary >= pos; {
		if (pos & sizeOfDictionary) != 0 {
			numBits = shifts + uint(1)
		}
		shifts++
		pos = 1 << shifts
	}

	mask := int64(1<<numBits - 1)

	return &nameGenerator{
		dictionary:           dictionary,
		numBits:              numBits,
		mask:                 mask,
		maxDictionaryIndices: 63 / numBits,
		randGen:              rand.NewSource(time.Now().UnixNano()),
	}
}

func (ng *nameGenerator) Generate(n int) string {
	b := make([]byte, n)
	for i, cache, numOfIndicesRemained := n-1, ng.nextRand(), ng.maxDictionaryIndices; i >= 0; {
		if numOfIndicesRemained == 0 {
			cache, numOfIndicesRemained = ng.nextRand(), ng.maxDictionaryIndices
		}
		if idx := int(cache & ng.mask); idx < len(ng.dictionary) {
			b[i] = ng.dictionary[idx]
			i--
		}
		cache >>= ng.numBits
		numOfIndicesRemained--
	}
	return string(b)
}

func (ng *nameGenerator) nextRand() int64 {
	ng.Lock()
	defer ng.Unlock()
	return ng.randGen.Int63()
}
