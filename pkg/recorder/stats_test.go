package recorder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAverage(t *testing.T) {
	t.Parallel()
	stats := []*Stat{
		{Bandwidth: 1.0, Latency: time.Second, Rtt: time.Second, Seq: 1},
		{Bandwidth: 2.0, Latency: 2 * time.Second, Rtt: 2 * time.Second, Seq: 2},
		{Bandwidth: 3.0, Latency: 3 * time.Second, Rtt: 3 * time.Second, Seq: 3},
		{Bandwidth: 4.0, Latency: 4 * time.Second, Rtt: 4 * time.Second, Seq: 4},
		{Bandwidth: 5.0, Latency: 5 * time.Second, Rtt: 5 * time.Second, Seq: 5},
	}

	avg := Average(stats)
	assert.Equal(t, 3.0, avg.Bandwidth)
	assert.Equal(t, 3*time.Second, avg.Latency)
	assert.Equal(t, 3*time.Second, avg.Rtt)
}

func TestIQRFilter(t *testing.T) {
	t.Parallel()

	t.Run("no outliers", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{
			{Bandwidth: 1.0},
			{Bandwidth: 2.0},
			{Bandwidth: 3.0},
			{Bandwidth: 4.0},
			{Bandwidth: 5.0},
		}

		filtered, err := IQRFilter(stats)
		assert.NoError(t, err)
		assert.ElementsMatch(t, stats, filtered)
	})

	t.Run("outliers", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{
			{Bandwidth: 1.0},
			{Bandwidth: 31.0},
			{Bandwidth: 30.0},
			{Bandwidth: 35.0},
			{Bandwidth: 29.0},
			{Bandwidth: 100.0},
		}

		filtered, err := IQRFilter(stats)
		assert.NoError(t, err)
		assert.Len(t, filtered, len(stats)-2)
		assert.NotContains(t, filtered, &Stat{Bandwidth: 1.0})
		assert.NotContains(t, filtered, &Stat{Bandwidth: 100.0})
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{}
		_, err := IQRFilter(stats)
		if ok := assert.Error(t, err); ok {
			assert.EqualError(t, err, "Input must not be empty.")
		}
	})
}

func TestMADFilter(t *testing.T) {
	t.Parallel()

	t.Run("no outliers", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{
			{Bandwidth: 1.0},
			{Bandwidth: 2.0},
			{Bandwidth: 3.0},
			{Bandwidth: 4.0},
			{Bandwidth: 5.0},
		}

		filtered, err := MADFilter(stats)
		assert.NoError(t, err)
		assert.ElementsMatch(t, stats, filtered)
	})

	t.Run("outliers", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{
			{Bandwidth: 1.0},
			{Bandwidth: 31.0},
			{Bandwidth: 30.0},
			{Bandwidth: 35.0},
			{Bandwidth: 29.0},
			{Bandwidth: 100.0},
		}

		filtered, err := MADFilter(stats)
		assert.NoError(t, err)
		assert.Len(t, filtered, len(stats)-2)
		assert.NotContains(t, filtered, &Stat{Bandwidth: 1.0})
		assert.NotContains(t, filtered, &Stat{Bandwidth: 100.0})
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{}
		_, err := MADFilter(stats)
		if ok := assert.Error(t, err); ok {
			assert.EqualError(t, err, "Input must not be empty.")
		}
	})
}

func TestSanitize(t *testing.T) {
	t.Parallel()
	stats := []*Stat{
		{Bandwidth: 1.0},
		{Bandwidth: 0.0},
		{Bandwidth: 3.0},
		{Bandwidth: -1.0},
		{Bandwidth: 5.0},
	}

	sanitized := Sanitize(stats)
	assert.Len(t, sanitized, 4)
	assert.NotContains(t, sanitized, &Stat{Bandwidth: -1.0})
}

func TestProcessStats(t *testing.T) {
	t.Parallel()
	t.Run("no stats", func(t *testing.T) {
		t.Parallel()
		stats := []*Stat{}
		_, err := ProcessStats(stats)
		if ok := assert.Error(t, err); ok {
			assert.EqualError(t, err, "Input must not be empty.")
		}
	})

	t.Run("stats", func(t *testing.T) {
		stats := []*Stat{
			{Bandwidth: -30.0},
			{Bandwidth: 1.0},
			{Bandwidth: 100.0},
			{Bandwidth: 30.0, Latency: 1 * time.Second, Rtt: 1 * time.Second, Seq: 1},
			{Bandwidth: 31.0, Latency: 2 * time.Second, Rtt: 2 * time.Second, Seq: 2},
			{Bandwidth: 29.0, Latency: 3 * time.Second, Rtt: 3 * time.Second, Seq: 3},
			{Bandwidth: 32.0, Latency: 4 * time.Second, Rtt: 4 * time.Second, Seq: 3},
			{Bandwidth: 33.0, Latency: 5 * time.Second, Rtt: 5 * time.Second, Seq: 3},
		}

		avg, err := ProcessStats(stats)
		assert.NoError(t, err)
		assert.Equal(t, 31.0, avg.Bandwidth)
		assert.Equal(t, 3*time.Second, avg.Latency)
		assert.Equal(t, 3*time.Second, avg.Rtt)
	})
}
