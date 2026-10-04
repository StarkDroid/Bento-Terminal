package build

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestInspectFixtures(t *testing.T) {
	root := filepath.Join("testdata")
	cases := []struct {
		dir     string
		systems []string
		auto    []string
		manual  []string
	}{
		{
			dir:     "npm",
			systems: []string{"npm"},
			auto:    []string{"npm run build", "npm run dev", "npm run test"},
			manual:  []string{"npm test", "npm run build"},
		},
		{
			dir:     "make",
			systems: []string{"Make"},
			auto:    []string{"make build", "make test", "make deploy"},
			manual:  []string{"make", "make test", "make build"},
		},
		{
			dir:     "gomod",
			systems: []string{"Go"},
			auto:    []string{"go test ./...", "go build ./...", "go vet ./..."},
			manual:  []string{"go test ./...", "go build ./..."},
		},
		{
			dir:     "cargo",
			systems: []string{"Cargo"},
			auto:    []string{"cargo test", "cargo build", "cargo check", "cargo run"},
			manual:  []string{"cargo test", "cargo build"},
		},
		{
			dir:     "py",
			systems: []string{"Python"},
			auto:    []string{"python -m pytest", "python -m build"},
			manual:  []string{"python -m pytest"},
		},
		{
			dir:     "gradle",
			systems: []string{"Gradle"},
			auto:    []string{"./gradlew hello", "./gradlew integrationTest"},
			manual:  []string{"./gradlew test", "./gradlew build", "./gradlew check"},
		},
		{
			dir:     "empty",
			systems: nil,
			auto:    nil,
			manual:  nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			got := Inspect(filepath.Join(root, tc.dir))
			if !reflect.DeepEqual(got.Systems, tc.systems) {
				t.Fatalf("systems = %#v", got.Systems)
			}
			if !reflect.DeepEqual(names(got.Auto), tc.auto) {
				t.Fatalf("auto = %#v", names(got.Auto))
			}
			if !reflect.DeepEqual(names(got.Manual), tc.manual) {
				t.Fatalf("manual = %#v", names(got.Manual))
			}
			for _, c := range append(append([]Command{}, got.Auto...), got.Manual...) {
				if len(c.Argv) == 0 {
					t.Fatalf("empty argv for %s", c.Name)
				}
				if c.Argv[0] == "sh" || c.Argv[0] == "bash" {
					t.Fatalf("command %s uses a shell: %#v", c.Name, c.Argv)
				}
			}
		})
	}
}

func TestNpmRunArgv(t *testing.T) {
	got := Inspect(filepath.Join("testdata", "npm"))
	var test Command
	for _, c := range got.Auto {
		if c.Name == "npm run test" {
			test = c
		}
	}
	want := []string{"npm", "run", "test"}
	if !reflect.DeepEqual(test.Argv, want) {
		t.Fatalf("argv = %#v", test.Argv)
	}
}

func names(cmds []Command) []string {
	if len(cmds) == 0 {
		return nil
	}
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.Name
	}
	return out
}
