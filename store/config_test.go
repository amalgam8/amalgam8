package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigMinimumEqualMaximum(t *testing.T) {

	def := time.Duration(15) * time.Second
	min := time.Duration(15) * time.Second
	max := time.Duration(15) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)

	assert.NotNil(t, c)
	assert.Equal(t, def, c.DefaultTTL)
	assert.Equal(t, min, c.MinimumTTL)
	assert.Equal(t, max, c.MaximumTTL)

}

func TestConfigMinimumSmallerThanMaximum(t *testing.T) {

	def := time.Duration(15) * time.Second
	min := time.Duration(10) * time.Second
	max := time.Duration(20) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)

	assert.NotNil(t, c)
	assert.Equal(t, def, c.DefaultTTL)
	assert.Equal(t, min, c.MinimumTTL)
	assert.Equal(t, max, c.MaximumTTL)

}

func TestConfigMinimumLargerThanMaximum(t *testing.T) {

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	def := time.Duration(15) * time.Second
	min := time.Duration(20) * time.Second
	max := time.Duration(10) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)
	assert.Fail(t, "Expected panic but got %v", c)

}

func TestConfigDefaultIsMinimum(t *testing.T) {

	def := time.Duration(10) * time.Second
	min := time.Duration(10) * time.Second
	max := time.Duration(20) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)

	assert.NotNil(t, c)
	assert.Equal(t, def, c.DefaultTTL)
	assert.Equal(t, min, c.MinimumTTL)
	assert.Equal(t, max, c.MaximumTTL)

}

func TestConfigDefaultIsMaximum(t *testing.T) {

	def := time.Duration(20) * time.Second
	min := time.Duration(10) * time.Second
	max := time.Duration(20) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)

	assert.NotNil(t, c)
	assert.Equal(t, def, c.DefaultTTL)
	assert.Equal(t, min, c.MinimumTTL)
	assert.Equal(t, max, c.MaximumTTL)

}

func TestConfigDefaultSmallerThanMinimum(t *testing.T) {

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	def := time.Duration(5) * time.Second
	min := time.Duration(20) * time.Second
	max := time.Duration(10) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)
	assert.Fail(t, "Expected panic but got %v", c)

}

func TestConfigDefaultLargerThanMaximum(t *testing.T) {

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	def := time.Duration(25) * time.Second
	min := time.Duration(20) * time.Second
	max := time.Duration(10) * time.Second

	c := NewConfig(def, min, max, defaultNamespaceCapacity)
	assert.Fail(t, "Expected panic but got %v", c)

}

func TestConfigCapacityNotValid(t *testing.T) {

	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	c := NewConfig(defaultDefaultTTL, defaultMinimumTTL, defaultMaximumTTL, -2)

	assert.Fail(t, "Expected panic but got %v", c)
}
