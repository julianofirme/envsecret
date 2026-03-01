// White-box tests for unexported helpers in cmd/set.go.
// Must use package cmd (not cmd_test) because validateKey, parseEnvLine, and
// readEnvLines are unexported.
package cmd

import (
	"bufio"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// validateKey
// ---------------------------------------------------------------------------

func TestValidateKey(t *testing.T) {
	cases := []struct {
		key     string
		wantErr bool
	}{
		// valid
		{"FOO", false},
		{"_PRIVATE", false},
		{"FOO_BAR_123", false},
		{"A", false},
		{"_", false},
		{"Z9", false},
		// invalid: empty
		{"", true},
		// invalid: starts with digit
		{"1FOO", true},
		// invalid: lowercase
		{"foo", true},
		{"Foo", true},
		// invalid: hyphen
		{"FOO-BAR", true},
		// invalid: space
		{"FOO BAR", true},
		// invalid: dot
		{"FOO.BAR", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.key, func(t *testing.T) {
			err := validateKey(tc.key)
			if tc.wantErr && err == nil {
				t.Errorf("validateKey(%q): expected error, got nil", tc.key)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateKey(%q): unexpected error: %v", tc.key, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseEnvLine
// ---------------------------------------------------------------------------

func TestParseEnvLine(t *testing.T) {
	cases := []struct {
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		// plain KEY=VALUE
		{"FOO=bar", "FOO", "bar", true},
		// double-quoted value
		{`BAZ="hello world"`, "BAZ", "hello world", true},
		// single-quoted value
		{"QUX='single'", "QUX", "single", true},
		// export prefix
		{"export MY_VAR=value", "MY_VAR", "value", true},
		// export with double-quoted value
		{`export QUOTED="quoted value"`, "QUOTED", "quoted value", true},
		// value with equals sign inside
		{"DB_URL=postgres://host/db?ssl=true", "DB_URL", "postgres://host/db?ssl=true", true},
		// empty value
		{"EMPTY=", "EMPTY", "", true},
		// blank line
		{"", "", "", false},
		// whitespace only
		{"   ", "", "", false},
		// comment
		{"# this is a comment", "", "", false},
		// no equals sign
		{"NOEQUALS", "", "", false},
		// leading/trailing whitespace: whole line is TrimSpace'd first,
		// so trailing spaces on the value are also stripped.
		{"  SPACED = val  ", "SPACED", " val", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.line, func(t *testing.T) {
			k, v, ok := parseEnvLine(tc.line)
			if ok != tc.wantOK {
				t.Errorf("parseEnvLine(%q): ok=%v, want %v", tc.line, ok, tc.wantOK)
				return
			}
			if ok {
				if k != tc.wantKey {
					t.Errorf("parseEnvLine(%q): key=%q, want %q", tc.line, k, tc.wantKey)
				}
				if v != tc.wantValue {
					t.Errorf("parseEnvLine(%q): value=%q, want %q", tc.line, v, tc.wantValue)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// readEnvLines
// ---------------------------------------------------------------------------

func TestReadEnvLines(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  [][2]string
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "single pair",
			input: "FOO=bar\n",
			want:  [][2]string{{"FOO", "bar"}},
		},
		{
			name:  "multiple pairs",
			input: "FOO=bar\nBAZ=qux\n",
			want:  [][2]string{{"FOO", "bar"}, {"BAZ", "qux"}},
		},
		{
			name:  "skip comments and blanks",
			input: "# comment\n\nFOO=bar\n\n# another\nBAZ=qux\n",
			want:  [][2]string{{"FOO", "bar"}, {"BAZ", "qux"}},
		},
		{
			name:  "export prefix",
			input: "export MY_VAR=hello\n",
			want:  [][2]string{{"MY_VAR", "hello"}},
		},
		{
			name:  "no trailing newline",
			input: "FOO=bar",
			want:  [][2]string{{"FOO", "bar"}},
		},
		{
			name:  "quoted values",
			input: "A=\"double\"\nB='single'\n",
			want:  [][2]string{{"A", "double"}, {"B", "single"}},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			got, err := readEnvLines(r)
			if err != nil {
				t.Fatalf("readEnvLines: unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len(got)=%d, len(want)=%d\ngot:  %v\nwant: %v",
					len(got), len(tc.want), got, tc.want)
			}
			for i, pair := range tc.want {
				if got[i] != pair {
					t.Errorf("pair[%d]: got %v, want %v", i, got[i], pair)
				}
			}
		})
	}
}
