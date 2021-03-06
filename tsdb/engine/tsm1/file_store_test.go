package tsm1_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
)

func TestFileStore_Read(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	values, err := fs.Read("cpu", 1)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[1]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToAsc_FromStart(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[0]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToAsc_Duplicate(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 4.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[1].values[0],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	// Check that calling Next will dedupe points
	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[3].values[0],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	exp = nil
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToAsc_BeforeStart(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[0]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

// Tests that seeking and reading all blocks that contain overlapping points does
// not skip any blocks.
func TestFileStore_SeekToAsc_BeforeStart_OverlapFloat(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 0.0), tsm1.NewValue(1, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 4.0), tsm1.NewValue(2, 7.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[3].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[2].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

// Tests that seeking and reading all blocks that contain overlapping points does
// not skip any blocks.
func TestFileStore_SeekToAsc_BeforeStart_OverlapInteger(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, int64(0)), tsm1.NewValue(1, int64(1))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(2))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, int64(3))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, int64(4)), tsm1.NewValue(2, int64(7))}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.IntegerValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[3].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[2].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

// Tests that seeking and reading all blocks that contain overlapping points does
// not skip any blocks.
func TestFileStore_SeekToAsc_BeforeStart_OverlapBoolean(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, true), tsm1.NewValue(1, false)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, false), tsm1.NewValue(2, true)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.BooleanValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[3].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[2].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

// Tests that seeking and reading all blocks that contain overlapping points does
// not skip any blocks.
func TestFileStore_SeekToAsc_BeforeStart_OverlapString(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, "zero"), tsm1.NewValue(1, "one")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "two")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, "three")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, "four"), tsm1.NewValue(2, "seven")}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.StringValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[3].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[2].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

// Tests that blocks with a lower min time in later files are not returned
// more than once causing unsorted results
func TestFileStore_SeekToAsc_OverlapMinFloat(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0), tsm1.NewValue(3, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0), tsm1.NewValue(4, 4.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 0.0), tsm1.NewValue(1, 1.1)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.2)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
		data[3].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	// Check that calling Next will dedupe points
	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	exp = nil
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

// Tests that blocks with a lower min time in later files are not returned
// more than once causing unsorted results
func TestFileStore_SeekToAsc_OverlapMinInteger(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, int64(1)), tsm1.NewValue(3, int64(3))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(2)), tsm1.NewValue(4, int64(4))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, int64(0)), tsm1.NewValue(1, int64(10))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(5))}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.IntegerValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
		data[3].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	// Check that calling Next will dedupe points
	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	exp = nil
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

// Tests that blocks with a lower min time in later files are not returned
// more than once causing unsorted results
func TestFileStore_SeekToAsc_OverlapMinBoolean(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, true), tsm1.NewValue(3, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, true), tsm1.NewValue(4, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, true), tsm1.NewValue(1, false)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, false)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.BooleanValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
		data[3].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	// Check that calling Next will dedupe points
	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	exp = nil
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

// Tests that blocks with a lower min time in later files are not returned
// more than once causing unsorted results
func TestFileStore_SeekToAsc_OverlapMinString(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, "1.0"), tsm1.NewValue(3, "3.0")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "2.0"), tsm1.NewValue(4, "4.0")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, "0.0"), tsm1.NewValue(1, "1.1")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "2.2")}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.StringValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	// Search for an entry that exists in the second file
	values, err := c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
		data[3].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	// Check that calling Next will dedupe points
	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	exp = nil
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToAsc_Middle(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0),
			tsm1.NewValue(2, 2.0),
			tsm1.NewValue(3, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(4, 4.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 3, true)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{data[0].values[2]}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{data[1].values[0]}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

}

func TestFileStore_SeekToAsc_End(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 2, true)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[2]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_FromStart(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp := data[0]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_Duplicate(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 4.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 2, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp := []tsm1.Value{
		data[3].values[0],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp = []tsm1.Value{
		data[1].values[0],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_OverlapMaxFloat(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0), tsm1.NewValue(3, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0), tsm1.NewValue(4, 4.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 0.0), tsm1.NewValue(1, 1.1)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.2)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 5, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp = []tsm1.Value{

		data[2].values[0],
		data[2].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_OverlapMaxInteger(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, int64(1)), tsm1.NewValue(3, int64(3))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(2)), tsm1.NewValue(4, int64(4))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, int64(0)), tsm1.NewValue(1, int64(10))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(5))}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.IntegerValue, 1000)
	c := fs.KeyCursor("cpu", 5, false)
	values, err := c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp = []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_OverlapMaxBoolean(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, true), tsm1.NewValue(3, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, true), tsm1.NewValue(4, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, true), tsm1.NewValue(1, false)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, false)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.BooleanValue, 1000)
	c := fs.KeyCursor("cpu", 5, false)
	values, err := c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp = []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_OverlapMaxString(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, "1.0"), tsm1.NewValue(3, "3.0")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "2.0"), tsm1.NewValue(4, "4.0")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, "0.0"), tsm1.NewValue(1, "1.1")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "2.2")}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.StringValue, 1000)
	c := fs.KeyCursor("cpu", 5, false)
	values, err := c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[1],
		data[1].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}
	exp = []tsm1.Value{
		data[2].values[0],
		data[2].values[1],
	}
	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_AfterEnd(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 4, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[2]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestFileStore_SeekToDesc_AfterEnd_OverlapFloat(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 4 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(8, 0.0), tsm1.NewValue(9, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, 3.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, 4.0), tsm1.NewValue(7, 7.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 10, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[0].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[3].values[0],
		data[3].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	if got, exp := len(values), 0; got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToDesc_AfterEnd_OverlapInteger(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(8, int64(0)), tsm1.NewValue(9, int64(1))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, int64(2))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, int64(3))}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, int64(4)), tsm1.NewValue(10, int64(7))}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.IntegerValue, 1000)
	c := fs.KeyCursor("cpu", 11, false)
	values, err := c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[3].values[0],
		data[0].values[0],
		data[0].values[1],
		data[3].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	if got, exp := len(values), 0; got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToDesc_AfterEnd_OverlapBoolean(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(8, true), tsm1.NewValue(9, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, true)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, false)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, true), tsm1.NewValue(7, false)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.BooleanValue, 1000)
	c := fs.KeyCursor("cpu", 11, false)
	values, err := c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[0].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[3].values[0],
		data[3].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	if got, exp := len(values), 0; got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToDesc_AfterEnd_OverlapString(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(8, "eight"), tsm1.NewValue(9, "nine")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, "two")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, "three")}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(3, "four"), tsm1.NewValue(7, "seven")}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.StringValue, 1000)
	c := fs.KeyCursor("cpu", 11, false)
	values, err := c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[0].values[0],
		data[0].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[3].values[0],
		data[3].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[1].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	if got, exp := len(values), 0; got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToDesc_Middle(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 1.0)}},
		keyValues{"cpu", []tsm1.Value{
			tsm1.NewValue(2, 2.0),
			tsm1.NewValue(3, 3.0),
			tsm1.NewValue(4, 4.0)},
		},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Search for an entry that exists in the second file
	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 3, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := []tsm1.Value{
		data[1].values[0],
		data[1].values[1],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp = []tsm1.Value{
		data[0].values[0],
	}

	if got, exp := len(values), len(exp); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}

	c.Next()
	values, err = c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)

	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	if got, exp := len(values), 0; got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_SeekToDesc_End(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 2, false)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	exp := data[2]
	if got, exp := len(values), len(exp.values); got != exp {
		t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
	}

	for i, v := range exp.values {
		if got, exp := values[i].Value(), v.Value(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", i, got, exp)
		}
	}
}

func TestKeyCursor_TombstoneRange(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	expValues := []int{0, 2}
	for _, v := range expValues {
		values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
		if err != nil {
			t.Fatalf("unexpected error reading values: %v", err)
		}

		exp := data[v]
		if got, exp := len(values), 1; got != exp {
			t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := values[0].String(), exp.values[0].String(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", 0, got, exp)
		}
		c.Next()
	}
}

func TestKeyCursor_TombstoneRange_PartialFloat(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{
			tsm1.NewValue(0, 1.0),
			tsm1.NewValue(1, 2.0),
			tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	buf := make([]tsm1.FloatValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadFloatBlock(&tsm1.TimeDecoder{}, &tsm1.FloatDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	expValues := []tsm1.Value{data[0].values[0], data[0].values[2]}
	for i, v := range expValues {
		exp := v
		if got, exp := len(values), 2; got != exp {
			t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := values[i].String(), exp.String(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", 0, got, exp)
		}
	}
}

func TestKeyCursor_TombstoneRange_PartialInteger(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{
			tsm1.NewValue(0, int64(1)),
			tsm1.NewValue(1, int64(2)),
			tsm1.NewValue(2, int64(3))}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	buf := make([]tsm1.IntegerValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadIntegerBlock(&tsm1.TimeDecoder{}, &tsm1.IntegerDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	expValues := []tsm1.Value{data[0].values[0], data[0].values[2]}
	for i, v := range expValues {
		exp := v
		if got, exp := len(values), 2; got != exp {
			t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := values[i].String(), exp.String(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", 0, got, exp)
		}
	}
}

func TestKeyCursor_TombstoneRange_PartialString(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{
			tsm1.NewValue(0, "1"),
			tsm1.NewValue(1, "2"),
			tsm1.NewValue(2, "3")}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	buf := make([]tsm1.StringValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadStringBlock(&tsm1.TimeDecoder{}, &tsm1.StringDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	expValues := []tsm1.Value{data[0].values[0], data[0].values[2]}
	for i, v := range expValues {
		exp := v
		if got, exp := len(values), 2; got != exp {
			t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := values[i].String(), exp.String(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", 0, got, exp)
		}
	}
}

func TestKeyCursor_TombstoneRange_PartialBoolean(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{
			tsm1.NewValue(0, true),
			tsm1.NewValue(1, false),
			tsm1.NewValue(2, true)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	buf := make([]tsm1.BooleanValue, 1000)
	c := fs.KeyCursor("cpu", 0, true)
	values, err := c.ReadBooleanBlock(&tsm1.TimeDecoder{}, &tsm1.BooleanDecoder{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error reading values: %v", err)
	}

	expValues := []tsm1.Value{data[0].values[0], data[0].values[2]}
	for i, v := range expValues {
		exp := v
		if got, exp := len(values), 2; got != exp {
			t.Fatalf("value length mismatch: got %v, exp %v", got, exp)
		}

		if got, exp := values[i].String(), exp.String(); got != exp {
			t.Fatalf("read value mismatch(%d): got %v, exp %v", 0, got, exp)
		}
	}
}

func TestFileStore_Open(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)

	// Create 3 TSM files...
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	_, err := newFileDir(dir, data...)
	if err != nil {
		fatal(t, "creating test files", err)
	}

	fs := tsm1.NewFileStore(dir)
	if err := fs.Open(); err != nil {
		fatal(t, "opening file store", err)
	}
	defer fs.Close()

	if got, exp := fs.Count(), 3; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}

	if got, exp := fs.CurrentGeneration(), 4; got != exp {
		t.Fatalf("current ID mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_Remove(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)

	// Create 3 TSM files...
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	files, err := newFileDir(dir, data...)
	if err != nil {
		fatal(t, "creating test files", err)
	}

	fs := tsm1.NewFileStore(dir)
	if err := fs.Open(); err != nil {
		fatal(t, "opening file store", err)
	}
	defer fs.Close()

	if got, exp := fs.Count(), 3; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}

	if got, exp := fs.CurrentGeneration(), 4; got != exp {
		t.Fatalf("current ID mismatch: got %v, exp %v", got, exp)
	}

	fs.Remove(files[2])

	if got, exp := fs.Count(), 2; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}

	if got, exp := fs.CurrentGeneration(), 4; got != exp {
		t.Fatalf("current ID mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_Open_Deleted(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)

	// Create 3 TSM files...
	data := []keyValues{
		keyValues{"cpu,host=server2!~#!value", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu,host=server1!~#!value", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem,host=server1!~#!value", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	_, err := newFileDir(dir, data...)
	if err != nil {
		fatal(t, "creating test files", err)
	}

	fs := tsm1.NewFileStore(dir)
	if err := fs.Open(); err != nil {
		fatal(t, "opening file store", err)
	}
	defer fs.Close()

	if got, exp := len(fs.Keys()), 3; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}

	if err := fs.Delete([]string{"cpu,host=server2!~#!value"}); err != nil {
		fatal(t, "deleting", err)
	}

	fs2 := tsm1.NewFileStore(dir)
	if err := fs2.Open(); err != nil {
		fatal(t, "opening file store", err)
	}
	defer fs2.Close()

	if got, exp := len(fs2.Keys()), 2; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_Delete(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu,host=server2!~#!value", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu,host=server1!~#!value", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem,host=server1!~#!value", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	keys := fs.Keys()
	if got, exp := len(keys), 3; got != exp {
		t.Fatalf("key length mismatch: got %v, exp %v", got, exp)
	}

	if err := fs.Delete([]string{"cpu,host=server2!~#!value"}); err != nil {
		fatal(t, "deleting", err)
	}

	keys = fs.Keys()
	if got, exp := len(keys), 2; got != exp {
		t.Fatalf("key length mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_Stats(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)

	// Create 3 TSM files...
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"mem", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
	}

	_, err := newFileDir(dir, data...)
	if err != nil {
		fatal(t, "creating test files", err)
	}

	fs := tsm1.NewFileStore(dir)
	if err := fs.Open(); err != nil {
		fatal(t, "opening file store", err)
	}
	defer fs.Close()

	stats := fs.Stats()
	if got, exp := len(stats), 3; got != exp {
		t.Fatalf("file count mismatch: got %v, exp %v", got, exp)
	}
}

func TestFileStore_CreateSnapshot(t *testing.T) {
	dir := MustTempDir()
	defer os.RemoveAll(dir)
	fs := tsm1.NewFileStore(dir)

	// Setup 3 files
	data := []keyValues{
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(0, 1.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(1, 2.0)}},
		keyValues{"cpu", []tsm1.Value{tsm1.NewValue(2, 3.0)}},
	}

	files, err := newFiles(dir, data...)
	if err != nil {
		t.Fatalf("unexpected error creating files: %v", err)
	}

	fs.Add(files...)

	// Create a tombstone
	if err := fs.DeleteRange([]string{"cpu"}, 1, 1); err != nil {
		t.Fatalf("unexpected error delete range: %v", err)
	}

	s, e := fs.CreateSnapshot()
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("temp file for hard links: %q", s)

	tfs, e := ioutil.ReadDir(s)
	if e != nil {
		t.Fatal(e)
	}
	if len(tfs) == 0 {
		t.Fatal("no files found")
	}

	for _, f := range fs.Files() {
		p := filepath.Join(s, filepath.Base(f.Path()))
		t.Logf("checking for existence of hard link %q", p)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Fatalf("unable to find file %q", p)
		}
		for _, tf := range f.TombstoneFiles() {
			p := filepath.Join(s, filepath.Base(tf.Path))
			t.Logf("checking for existence of hard link %q", p)
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Fatalf("unable to find file %q", p)
			}
		}
	}
}

func newFileDir(dir string, values ...keyValues) ([]string, error) {
	var files []string

	id := 1
	for _, v := range values {
		f := MustTempFile(dir)
		w, err := tsm1.NewTSMWriter(f)
		if err != nil {
			return nil, err
		}

		if err := w.Write(v.key, v.values); err != nil {
			return nil, err
		}

		if err := w.WriteIndex(); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}
		newName := filepath.Join(filepath.Dir(f.Name()), tsmFileName(id))
		if err := os.Rename(f.Name(), newName); err != nil {
			return nil, err
		}
		id++

		files = append(files, newName)
	}
	return files, nil

}

func newFiles(dir string, values ...keyValues) ([]tsm1.TSMFile, error) {
	var files []tsm1.TSMFile

	id := 1
	for _, v := range values {
		f := MustTempFile(dir)
		w, err := tsm1.NewTSMWriter(f)
		if err != nil {
			return nil, err
		}

		if err := w.Write(v.key, v.values); err != nil {
			return nil, err
		}

		if err := w.WriteIndex(); err != nil {
			return nil, err
		}

		if err := w.Close(); err != nil {
			return nil, err
		}

		newName := filepath.Join(filepath.Dir(f.Name()), tsmFileName(id))
		if err := os.Rename(f.Name(), newName); err != nil {
			return nil, err
		}
		id++

		fd, err := os.Open(newName)
		if err != nil {
			return nil, err
		}
		r, err := tsm1.NewTSMReader(fd)
		if err != nil {
			return nil, err
		}
		files = append(files, r)
	}
	return files, nil
}

type keyValues struct {
	key    string
	values []tsm1.Value
}

func MustTempDir() string {
	dir, err := ioutil.TempDir("", "tsm1-test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp dir: %v", err))
	}
	return dir
}

func MustTempFile(dir string) *os.File {
	f, err := ioutil.TempFile(dir, "tsm1test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp file: %v", err))
	}
	return f
}

func fatal(t *testing.T, msg string, err error) {
	t.Fatalf("unexpected error %v: %v", msg, err)
}

func tsmFileName(id int) string {
	return fmt.Sprintf("%09d-%09d.tsm", id, 1)
}
