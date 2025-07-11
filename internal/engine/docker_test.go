package engine

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nalgeon/be"
	"github.com/nalgeon/codapi/internal/config"
	"github.com/nalgeon/codapi/internal/execy"
	"github.com/nalgeon/codapi/internal/logx"
)

var dockerCfg = &config.Config{
	Boxes: map[string]*config.Box{
		"alpine": {
			Image:   "codapi/alpine",
			Runtime: "runc",
			Host: config.Host{
				CPU: 1, Memory: 64, Network: "none",
				Volume: "%s:/sandbox:ro",
				NProc:  64,
			},
		},
		"go": {
			Image:   "codapi/go",
			Runtime: "runc",
			Host: config.Host{
				CPU: 1, Memory: 64, Network: "none",
				Volume: "%s:/sandbox:ro",
				NProc:  64,
			},
		},
		"go:dev": {
			Image:   "codapi/go:dev",
			Runtime: "runc",
			Host: config.Host{
				CPU: 1, Memory: 64, Network: "none",
				Volume: "%s:/sandbox:ro",
				NProc:  64,
			},
		},
		"python": {
			Image:   "codapi/python",
			Runtime: "runc",
			Host: config.Host{
				CPU: 1, Memory: 64, Network: "none",
				Volume: "%s:/sandbox:ro",
				NProc:  64,
			},
		},
		"python:dev": {
			Image:   "codapi/python:dev",
			Runtime: "runc",
			Host: config.Host{
				CPU: 1, Memory: 64, Network: "none",
				Volume: "%s:/sandbox:ro",
				NProc:  64,
			},
		},
	},
	Commands: map[string]config.SandboxCommands{
		"alpine": map[string]*config.Command{
			"echo": {
				Engine: "docker",
				Before: &config.Step{
					Box: "alpine", User: "sandbox", Action: "run", Detach: true,
					Command: []string{"echo", "before"},
					NOutput: 4096,
				},
				Steps: []*config.Step{
					{
						Box: ":name", User: "sandbox", Action: "exec",
						Command: []string{"sh", "main.sh"},
						NOutput: 4096,
					},
				},
				After: &config.Step{
					Box: ":name", User: "sandbox", Action: "stop",
					NOutput: 4096,
				},
			},
		},
		"go": map[string]*config.Command{
			"run": {
				Engine: "docker",
				Steps: []*config.Step{
					{
						Box: "go", User: "sandbox", Action: "run",
						Command: []string{"go", "build"},
						NOutput: 4096,
					},
					{
						Box: "alpine", Version: "latest",
						User: "sandbox", Action: "run",
						Command: []string{"./main"},
						NOutput: 4096,
					},
				},
			},
		},
		"postgresql": map[string]*config.Command{
			"run": {
				Engine: "docker",
				Before: &config.Step{
					Box: "postgres", User: "sandbox", Action: "exec",
					Command: []string{"psql", "-f", "create.sql"},
					NOutput: 4096,
				},
				Steps: []*config.Step{
					{
						Box: "postgres", User: "sandbox", Action: "exec", Stdin: true,
						Command: []string{"psql", "--user=:name"},
						NOutput: 4096,
					},
				},
				After: &config.Step{
					Box: "postgres", User: "sandbox", Action: "exec",
					Command: []string{"psql", "-f", "drop.sql"},
					NOutput: 4096,
				},
			},
		},
		"python": map[string]*config.Command{
			"run": {
				Engine: "docker",
				Entry:  "main.py",
				Steps: []*config.Step{
					{
						Box: "python", User: "sandbox", Action: "run",
						Command: []string{"python", "main.py"},
						NOutput: 4096,
					},
				},
			},
		},
	},
}

func TestDockerRun(t *testing.T) {
	logx.Mock()
	commands := map[string]execy.CmdOut{
		"docker run": {Stdout: "hello world", Stderr: "", Err: nil},
	}
	mem := execy.Mock(commands)

	t.Run("success", func(t *testing.T) {
		mem.Clear()
		engine := NewDocker(dockerCfg, "python", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "python",
			Command: "run",
			Files: map[string]string{
				"": "print('hello world')",
			},
		}
		out := engine.Exec(req)
		be.Equal(t, out.ID, req.ID)
		be.True(t, out.OK)
		want := "hello world"
		be.Equal(t, out.Stdout, want)
		be.Equal(t, out.Stderr, "")
		be.Equal(t, out.Err, nil)
		mem.MustHave(t, "codapi/python")
		mem.MustHave(t, "python main.py")
	})

	t.Run("latest version", func(t *testing.T) {
		mem.Clear()
		engine := NewDocker(dockerCfg, "python", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "python",
			Command: "run",
			Files: map[string]string{
				"": "print('hello world')",
			},
		}
		out := engine.Exec(req)
		be.True(t, out.OK)
		mem.MustHave(t, "codapi/python")
	})

	t.Run("custom version", func(t *testing.T) {
		mem.Clear()
		engine := NewDocker(dockerCfg, "python", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "python",
			Version: "dev",
			Command: "run",
			Files: map[string]string{
				"": "print('hello world')",
			},
		}
		out := engine.Exec(req)
		be.True(t, out.OK)
		mem.MustHave(t, "codapi/python:dev")
	})

	t.Run("step version", func(t *testing.T) {
		mem.Clear()
		engine := NewDocker(dockerCfg, "go", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "go",
			Version: "dev",
			Command: "run",
			Files: map[string]string{
				"": "var n = 42",
			},
		}
		out := engine.Exec(req)
		be.True(t, out.OK)
		mem.MustHave(t, "codapi/go:dev")
		mem.MustHave(t, "codapi/alpine")
	})

	t.Run("unsupported version", func(t *testing.T) {
		mem.Clear()
		engine := NewDocker(dockerCfg, "python", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "python",
			Version: "42",
			Command: "run",
			Files: map[string]string{
				"": "print('hello world')",
			},
		}
		out := engine.Exec(req)
		be.Equal(t, out.OK, false)
		want := "unknown box python:42"
		be.Equal(t, out.Stderr, want)
	})

	t.Run("directory traversal attack", func(t *testing.T) {
		mem.Clear()
		const fileName = "../../opt/codapi/codapi"
		engine := NewDocker(dockerCfg, "python", "run")
		req := Request{
			ID:      "http_42",
			Sandbox: "python",
			Command: "run",
			Files: map[string]string{
				"":       "print('hello world')",
				fileName: "hehe",
			},
		}
		out := engine.Exec(req)
		be.Equal(t, out.OK, false)
		want := fmt.Sprintf("files[%s]: invalid name", fileName)
		be.Equal(t, out.Stderr, want)
	})
}

func TestDockerExec(t *testing.T) {
	logx.Mock()
	commands := map[string]execy.CmdOut{
		"docker exec": {Stdout: "hello world", Stderr: "", Err: nil},
	}
	mem := execy.Mock(commands)
	engine := NewDocker(dockerCfg, "postgresql", "run")

	t.Run("success", func(t *testing.T) {
		req := Request{
			ID:      "http_42",
			Sandbox: "postgresql",
			Command: "run",
			Files: map[string]string{
				"": "select 'hello world'",
			},
		}
		out := engine.Exec(req)
		be.Equal(t, out.ID, req.ID)
		be.True(t, out.OK)
		want := "hello world"
		be.Equal(t, out.Stdout, want)
		be.Equal(t, out.Stderr, "")
		be.Equal(t, out.Err, nil)
		mem.MustHave(t, "psql -f create.sql")
		mem.MustHave(t, "psql --user=http_42")
		mem.MustHave(t, "psql -f drop.sql")
	})
}

func TestDockerStop(t *testing.T) {
	logx.Mock()
	commands := map[string]execy.CmdOut{
		"docker run":  {Stdout: "c958ff2", Stderr: "", Err: nil},
		"docker exec": {Stdout: "hello", Stderr: "", Err: nil},
		"docker stop": {Stdout: "alpine_42", Stderr: "", Err: nil},
	}
	mem := execy.Mock(commands)
	engine := NewDocker(dockerCfg, "alpine", "echo")

	t.Run("success", func(t *testing.T) {
		req := Request{
			ID:      "alpine_42",
			Sandbox: "alpine",
			Command: "echo",
			Files: map[string]string{
				"": "echo hello",
			},
		}
		out := engine.Exec(req)
		be.Equal(t, out.ID, req.ID)
		be.True(t, out.OK)
		want := "hello"
		be.Equal(t, out.Stdout, want)
		be.Equal(t, out.Stderr, "")
		be.Equal(t, out.Err, nil)
		mem.MustHave(t, "docker run --rm --name alpine_42", "--detach")
		mem.MustHave(t, "docker exec --interactive --user sandbox alpine_42 sh main.sh")
		mem.MustHave(t, "docker stop alpine_42")
	})
}

func Test_expandVars(t *testing.T) {
	const name = "codapi_01"
	commands := map[string]string{
		"python main.py":             "python main.py",
		"sh create.sh :name":         "sh create.sh " + name,
		"sh copy.sh :name new-:name": "sh copy.sh " + name + " new-" + name,
	}
	for cmd, want := range commands {
		src := strings.Fields(cmd)
		exp := expandVars(src, name)
		got := strings.Join(exp, " ")
		be.Equal(t, got, want)
	}
}
