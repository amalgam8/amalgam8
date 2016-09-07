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

package cluster

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"os"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/rcrowley/go-metrics"
)

const (
	directoryPermissions = 0777 // rwxrwxrwx
	filePermissions      = 0666 // rw-rw-rw
)

const (
	fsReadErrorsMetricsName  = "cluster.volume.errors.read"
	fsWriteErrorsMetricsName = "cluster.volume.errors.write"
)

func newFilesystemBackend(dir string, unreadableTimeout time.Duration) (*filesystemBackend, error) {
	b := &filesystemBackend{
		dir:               dir,
		unreadableTimeout: unreadableTimeout,
		cache:             make(map[MemberID]*member),
		readErrors:        metrics.NewRegisteredMeter(fsReadErrorsMetricsName, metrics.DefaultRegistry),
		writeErrors:       metrics.NewRegisteredMeter(fsWriteErrorsMetricsName, metrics.DefaultRegistry),
		logger:            logging.GetLogger(module),
	}

	err := b.initializeDirectory()
	if err != nil {
		return nil, err
	}

	return b, nil
}

type filesystemBackend struct {
	dir               string
	unreadableTimeout time.Duration
	cache             map[MemberID]*member
	readErrors        metrics.Meter
	writeErrors       metrics.Meter
	logger            *logrus.Entry
}

func (b *filesystemBackend) WriteMember(m *member) error {
	path := filepath.Join(b.dir, string(m.ID()))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermissions)

	if err != nil {
		b.readErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error opening member %v file at path %v", m.ID(), path)
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			b.writeErrors.Mark(1)
			b.logger.WithField("error", err).Warningf("Error closing member %v file at path %v", m.ID(), path)
		}
		return
	}()

	err = json.NewEncoder(file).Encode(m)

	if err != nil {
		b.writeErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error writing member %v file at path %v", m.ID(), path)
	} else {
		// Commits the current contents of the file to stable storage.
		err = file.Sync()
		if err != nil {
			b.writeErrors.Mark(1)
			b.logger.WithField("error", err).Warningf("Error flushing member %v file at path %v", m.ID(), path)
		}
	}

	return err
}

func (b *filesystemBackend) DeleteMember(id MemberID) error {
	path := filepath.Join(b.dir, string(id))
	err := os.Remove(path)

	if err != nil {
		b.writeErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error deleting member %v file at path %v", id, path)
	}

	return err
}

func (b *filesystemBackend) ReadMember(id MemberID) (*member, error) {
	var m *member
	path := filepath.Join(b.dir, string(id))
	file, err := os.Open(path)

	if err != nil {
		b.readErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error opening member %v file at path %v", id, path)
		return nil, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			b.writeErrors.Mark(1)
			b.logger.WithField("error", err).Warningf("Error closing member %v file at path %v", m.ID(), path)
		}
		return
	}()

	m = new(member)
	err = json.NewDecoder(file).Decode(m)
	if err == nil {
		return m, nil
	}

	b.readErrors.Mark(1)
	b.logger.WithField("error", err).Warningf("Error reading member %v file at path %v", id, path)

	if mcache, exists := b.cache[id]; exists {
		*m = *mcache
		err = nil
	} else {
		m = nil
	}

	// Attempt to delete unreadable files which are unmodified for a "long" time
	// Note: deletion should succeed despite the file is open
	stat, err2 := file.Stat()
	if err2 != nil {
		b.readErrors.Mark(1)
		b.logger.WithField("error", err2).Warningf("Error reading member %v file info at path %v", id, path)
	} else {
		unmodTime := time.Now().Sub(stat.ModTime())
		if unmodTime >= b.unreadableTimeout {
			b.logger.Warningf("Attempting to delete unreadble member %v file at path %v", id, path)
			err2 = os.Remove(path)
			if err2 != nil {
				b.writeErrors.Mark(1)
				b.logger.WithField("error", err2).Warningf("Error deleting unreadable member %v file at path %v", id, path)
			}
			return nil, err
		} else if m != nil && stat.ModTime().After(m.Timestamp) {
			m.Timestamp = stat.ModTime()
		}
	}

	return m, err
}

func (b *filesystemBackend) ReadMembers() (map[MemberID]*member, error) {
	ids, err := b.ReadMemberIDs()
	if err != nil {
		// error already logged by ReadMemberIDs()
		// if we couldn't read the directory we return the previous list
		return b.copyCache(), nil
	}

	// remove deleted members from the cache
	for id := range b.cache {
		if _, exists := ids[id]; !exists {
			delete(b.cache, id)
		}
	}

	for id := range ids {
		member, err := b.ReadMember(id)
		if err != nil {
			// error already logged by ReadMember(),
			// continue reading other members...
			delete(b.cache, id)
			continue
		}
		b.cache[id] = member
	}

	return b.copyCache(), nil
}

func (b *filesystemBackend) ReadMemberIDs() (map[MemberID]struct{}, error) {
	dir, err := os.Open(b.dir)
	if err != nil {
		b.readErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error opening cluster directory at path %v", dir)
		return nil, err
	}

	defer func() {
		err := dir.Close()
		if err != nil {
			b.writeErrors.Mark(1)
			b.logger.WithField("error", err).Warningf("Error closing cluster directory at path %v", dir)
		}
		return
	}()

	// File names are relative
	filenames, err := dir.Readdirnames(-1)
	if err != nil {
		b.readErrors.Mark(1)
		b.logger.WithField("error", err).Warningf("Error reading cluster directory at path %v", dir)
		return nil, err
	}

	ids := make(map[MemberID]struct{})
	for _, filename := range filenames {
		ids[MemberID(filename)] = struct{}{}
	}

	return ids, nil
}

func (b *filesystemBackend) initializeDirectory() error {
	stat, err := os.Stat(b.dir)

	// Something exists
	if stat != nil {
		if !stat.IsDir() {
			err := fmt.Errorf("Not a directory: %v", b.dir)
			b.logger.WithField("error", err).Error("Error initializating cluster directory")
			return err
		}
		// TODO: Fail-fast if directory is accessible (permissions-wise)
		return nil
	}

	if !os.IsNotExist(err) {
		b.logger.WithField("error", err).Error("Error initializating cluster directory")
		return err
	}

	// Create directory with global RW permissions
	err = os.MkdirAll(b.dir, os.ModeDir|directoryPermissions)
	if err != nil {
		b.logger.WithField("error", err).Warningf("Error initializating cluster directory")
	}
	return err
}

func (b *filesystemBackend) copyCache() map[MemberID]*member {
	newCache := make(map[MemberID]*member)
	for k, v := range b.cache {
		newCache[k] = v
	}
	return newCache
}
