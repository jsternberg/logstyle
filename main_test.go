package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

func TestZapLinter(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	dir, err := ioutil.TempDir(cwd, "logstyle-test")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer os.RemoveAll(dir)

	for _, tt := range []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "StringConstant",
			in: `package main
import "go.uber.org/zap"
func main() {
	logger := zap.NewNop()
	logger.Info("Hello, World!")
}
`,
		},
		{
			name: "Sprintf",
			in: `package main
import (
	"fmt"
	"go.uber.org/zap"
)
func main() {
	logger := zap.NewNop()
	logger.Info(fmt.Sprintf("Hello, %s!", "World"))
}
`,
			out: `main.go:8:2: call must use a string literal or a constant
`,
		},
		{
			name: "Struct",
			in: `package main
import "go.uber.org/zap"
type A struct {
	Logger *zap.Logger
}
func main() {
	a := A{
		Logger: zap.NewNop(),
	}
	msg := "Hello, World!"
	a.Logger.Info(msg)
}
`,
			out: `main.go:11:2: call must use a string literal or a constant
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := ioutil.TempDir(dir, "fakepkg")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer os.RemoveAll(dir)

			ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(tt.in), 0600)

			path, err := filepath.Rel(cwd, dir)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			var got bytes.Buffer
			if err := Analyze(&got, "./"+path); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if exp, got := tt.out, normalize(&got, path+"/"); exp != got {
				t.Fatalf("unexpected output:\n%s", diff.LineDiff(exp, got))
			}
		})
	}
}

func normalize(r io.Reader, prefix string) string {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Fprintln(&buf, strings.TrimPrefix(scanner.Text(), prefix))
	}
	return buf.String()
}
