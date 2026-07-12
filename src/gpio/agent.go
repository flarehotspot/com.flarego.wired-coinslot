package gpio

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

// coinAgentScript (the Python source) is defined in coin_agent_script.go as a Go
// string literal rather than //go:embed of a .py — see that file for why.

const (
	restartBackoffMin = 1 * time.Second
	restartBackoffMax = 5 * time.Second
	pulseBufferSize   = 64
)

// Logger is the minimal logging surface the agent needs; api.Logger() satisfies it.
type Logger interface {
	Info(message string) error
	Error(message string) error
}

// Config carries one coinslot's resolved hardware settings. For the Python
// drivers it is serialized to JSON and handed to coin_agent.py as argv[1]; the
// gpiod driver reads it directly in-process. Pin fields apply to rpi/opi
// (BOARD numbering); line fields apply to gpiod (Allwinner port names).
type Config struct {
	Driver  string `json:"driver"`  // "rpi" | "opi" | "gpiod" (Go-side selection)
	Library string `json:"library"` // "rpi" | "opi" (what coin_agent.py imports)
	Board   string `json:"board"`   // OPi.GPIO board module (opi only)

	// Physical header (BOARD) pin numbers — the operator-facing addressing for
	// every driver. rpi/opi pass these straight through; gpiod translates them
	// to char-device line offsets via Header below.
	CoinPin  int `json:"coin_pin"`
	RelayPin int `json:"relay_pin"`

	// gpiod only: the char-device chip (matched by pinctrl label), the board's
	// physical-pin -> line-name map, and the scheme for turning a line name into
	// a char-device offset ("sunxi" port name like "PH5", or "bcm" name like
	// "GPIO17"). Used to resolve CoinPin/RelayPin to line offsets. Not serialized
	// — gpiod never goes through Python.
	ChipLabel string         `json:"-"`
	Scheme    string         `json:"-"`
	Header    map[int]string `json:"-"`

	// Shared signal/relay settings.
	Pull        string `json:"pull"`
	Edge        string `json:"edge"`
	DebounceMs  int    `json:"debounce_ms"`
	RelayActive int    `json:"relay_active"`
}

// CoinAgent is the hardware surface the Manager drives, regardless of the
// underlying GPIO mechanism. Both the Python supervisor (*Agent) and the
// pure-Go char-device agent (*CdevAgent) implement it, so the Manager is
// driver-agnostic.
type CoinAgent interface {
	Pulses() <-chan struct{}
	Start() error
	Stop()
	OpenRelay()
	CloseRelay()
}

// NewCoinAgent builds the right agent for the resolved driver. gpiod runs
// in-process (no Python); everything else falls back to the supervised
// coin_agent.py.
func NewCoinAgent(cfg Config, logger Logger) CoinAgent {
	if cfg.Driver == "gpiod" {
		return NewCdevAgent(cfg, logger)
	}
	return NewAgent(cfg, logger)
}

// Compile-time guarantee that the Python supervisor satisfies the interface.
var _ CoinAgent = (*Agent)(nil)

type event struct {
	Event string `json:"event"`
	Msg   string `json:"msg"`
}

// Agent supervises one persistent coin_agent.py subprocess: it surfaces coin
// pulses on a channel, forwards relay commands over stdin, and transparently
// restarts the process (re-applying the desired relay state) if it dies.
type Agent struct {
	cfg        Config
	logger     Logger
	scriptPath string
	pulses     chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	mu        sync.Mutex
	stdin     io.Writer
	relayOpen bool
	started   bool
}

func NewAgent(cfg Config, logger Logger) *Agent {
	return &Agent{
		cfg:    cfg,
		logger: logger,
		pulses: make(chan struct{}, pulseBufferSize),
	}
}

// Pulses delivers one value per coin-acceptor pulse detected on the coin pin.
func (a *Agent) Pulses() <-chan struct{} { return a.pulses }

// Start materializes the embedded script and launches the supervised process.
func (a *Agent) Start() error {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return nil
	}
	a.started = true
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.mu.Unlock()

	scriptPath, err := materializeScript()
	if err != nil {
		return err
	}
	a.scriptPath = scriptPath

	go a.supervise()
	return nil
}

// Stop terminates the subprocess and stops the supervisor. The agent's stdin is
// closed by killing the process, which the script treats as a cleanup signal.
func (a *Agent) Stop() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// OpenRelay energizes the relay so the acceptor takes coins.
func (a *Agent) OpenRelay() {
	a.mu.Lock()
	a.relayOpen = true
	a.mu.Unlock()
	a.writeCmd("relay open")
}

// CloseRelay de-energizes the relay so the acceptor rejects coins.
func (a *Agent) CloseRelay() {
	a.mu.Lock()
	a.relayOpen = false
	a.mu.Unlock()
	a.writeCmd("relay close")
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

func (a *Agent) supervise() {
	backoff := restartBackoffMin
	for {
		if a.ctx.Err() != nil {
			return
		}

		start := time.Now()
		if err := a.runOnce(); err != nil && a.ctx.Err() == nil {
			// No python3 interpreter present: retrying won't help until the
			// system_packages are installed and the plugin restarts. Log once
			// and stop the supervisor (the pulse counter still runs, so the
			// dev/mock pulse path keeps working).
			if errors.Is(err, exec.ErrNotFound) {
				a.logf("python3 not found; coin agent disabled (install system_packages)")
				return
			}
			// Exit code 2 is coin_agent.py's "could not initialize GPIO" signal
			// (missing RPi/OPi library, unknown board module, invalid/busy pin).
			// None of these resolve by retrying without a config change, which
			// triggers a fresh agent via Manager.Reload — so log once and stop
			// instead of looping every few seconds. The pulse counter keeps
			// running, so the dev/mock pulse path is unaffected.
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
				a.logf("coin agent GPIO setup failed; disabled until settings change (check the board's GPIO library is installed and the pins are valid)")
				return
			}
			a.logf("coin agent exited: " + err.Error())
		}

		if a.ctx.Err() != nil {
			return
		}

		// A process that ran for a while is a transient crash; reset backoff.
		if time.Since(start) > restartBackoffMax {
			backoff = restartBackoffMin
		}

		select {
		case <-a.ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < restartBackoffMax {
			backoff *= 2
			if backoff > restartBackoffMax {
				backoff = restartBackoffMax
			}
		}
	}
}

func (a *Agent) runOnce() error {
	cfgJSON, err := json.Marshal(a.cfg)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(a.ctx, "python3", a.scriptPath, string(cfgJSON))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	a.mu.Lock()
	a.stdin = stdin
	relayOpen := a.relayOpen
	a.mu.Unlock()

	// Re-apply the desired relay state on (re)start so a crash never silently
	// leaves the relay in the wrong position.
	if relayOpen {
		a.writeCmd("relay open")
	} else {
		a.writeCmd("relay close")
	}

	go a.drainStderr(stderr)
	a.readEvents(stdout) // blocks until the process closes stdout

	a.mu.Lock()
	a.stdin = nil
	a.mu.Unlock()

	return cmd.Wait()
}

func (a *Agent) readEvents(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var ev event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		switch ev.Event {
		case "pulse":
			select {
			case a.pulses <- struct{}{}:
			default: // drop if the consumer is momentarily behind
			}
		case "ready":
			a.logf("coin agent ready")
		case "error":
			a.logf("coin agent error: " + ev.Msg)
		}
	}
}

func (a *Agent) drainStderr(stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		a.logf("coin agent stderr: " + scanner.Text())
	}
}

func (a *Agent) writeCmd(cmd string) {
	a.mu.Lock()
	w := a.stdin
	a.mu.Unlock()
	if w == nil {
		return
	}
	if _, err := w.Write([]byte(cmd + "\n")); err != nil {
		a.logf("failed to send command to coin agent: " + err.Error())
	}
}

func (a *Agent) logf(msg string) {
	if a.logger != nil {
		_ = a.logger.Error("[wired-coinslot] " + msg)
	}
}

func materializeScript() (string, error) {
	dir := filepath.Join(os.TempDir(), "com.flarego.wired-coinslot")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, "coin_agent.py")
	if err := os.WriteFile(p, coinAgentScript, 0o755); err != nil {
		return "", err
	}
	return p, nil
}
