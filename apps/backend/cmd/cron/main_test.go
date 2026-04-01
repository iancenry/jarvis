package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunJobReturnsErrorForUnknownJob(t *testing.T) {
	err := runJob("missing-job")
	require.Error(t, err)
	require.EqualError(t, err, "job 'missing-job' not found")
}
