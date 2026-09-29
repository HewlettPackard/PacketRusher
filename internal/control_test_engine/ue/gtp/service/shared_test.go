/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupConcurrency(t *testing.T) {
	for value, want := range map[string]int{"": 32, "8": 8, "0": 32, "-1": 32, "many": 32} {
		t.Setenv("PR_SETUP_SLOTS", value)
		assert.Equal(t, want, setupConcurrency(), "PR_SETUP_SLOTS=%q", value)
	}
}
