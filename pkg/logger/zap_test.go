package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// resetOnce is a helper function to reset the sync.Once variable between tests.
// This is crucial for testing the singleton creation logic multiple times.
func resetOnce() {
	once = sync.Once{}
	instance = nil
}

// TestNewLogger_Singleton ensures that NewLogger creates only one instance of the logger.
func TestNewLogger_Singleton(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "test_singleton.log"
	defer os.Remove(logFile)

	// First call to create the instance
	logger1, err := NewLogger("info", logFile)
	require.NoError(t, err, "First call to NewLogger should not produce an error")
	require.NotNil(t, logger1, "First logger instance should not be nil")

	// Second call should return the same instance
	logger2, err := NewLogger("debug", "another_file.log") // Parameters should be ignored
	require.NoError(t, err, "Second call to NewLogger should not produce an error")
	require.NotNil(t, logger2, "Second logger instance should not be nil")

	// Check if both pointers point to the same instance
	assert.Same(t, logger1, logger2, "NewLogger should return the same instance on subsequent calls")

	// Verify that the initial configuration is used
	// We can't directly inspect the level, but we can infer it from the initial file name.
	logger1.Info("Testing singleton instance.")
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Testing singleton instance.")
}

// TestNewLogger_FileLogging tests if the logger correctly writes to a specified file.
func TestNewLogger_FileLogging(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "test_file_logging.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("debug", logFile)
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("This is a test log message.", "key", "value")
	err = logger.Sync()
	require.NoError(t, err)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	// Check for the core parts of the log message
	assert.Contains(t, string(content), "[INFO] This is a test log message.")
	// The original assertion `assert.Contains(t, string(content), "\"key\":\"value\"")` failed
	// because the ConsoleEncoder pretty-prints the JSON for structured fields, which includes
	// a space after the colon (e.g., `"key": "value"` instead of `"key":"value"`).
	// We update the assertion to match the actual format produced by the logger.
	assert.Contains(t, string(content), `"key": "value"`)
}

// TestNewLogger_ConsoleLogging tests if the logger correctly writes to stdout when no file is specified.
func TestNewLogger_ConsoleLogging(t *testing.T) {
	resetOnce()
	defer resetOnce()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create logger with empty file path to force console logging
	logger, err := NewLogger("info", "")
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Info("Logging to console")

	// Restore stdout and read from the pipe
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	assert.Contains(t, output, "[INFO] Logging to console")
}

func TestNewLogger_InvalidLogLevel(t *testing.T) {
	resetOnce()
	defer resetOnce()

	// --- Step 1: Capture stderr BEFORE logger creation ---
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// --- Step 2: Capture stdout BEFORE logger creation ---
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// --- Step 3: Create logger (now both stderr and stdout are intercepted) ---
	logger, err := NewLogger("invalidLevel", "")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// --- Step 4: Perform logging ---
	logger.Debug("This should not be logged")
	logger.Info("This should be logged")

	// --- Step 5: Close and restore outputs ---
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout

	// --- Step 6: Read and check captured stderr ---
	var errBuf bytes.Buffer
	io.Copy(&errBuf, rErr)
	assert.Contains(t, errBuf.String(), "Invalid log level: invalidLevel")

	// --- Step 7: Read and check captured stdout ---
	var outBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	output := outBuf.String()
	assert.NotContains(t, output, "This should not be logged")
	assert.Contains(t, output, "[INFO] This should be logged")
}

// TestLoggingLevels verifies that all log levels write correctly.
func TestLoggingLevels(t *testing.T) {
	resetOnce()
	defer resetOnce()

	logFile := "test_levels.log"
	defer os.Remove(logFile)

	// Use "debug" level to ensure all messages are captured
	logger, err := NewLogger("debug", logFile)
	require.NoError(t, err)

	// Log messages at different levels
	logger.Debug("debug message", "id", 1)
	logger.Info("info message", "id", 2)
	logger.Warn("warn message", "id", 3)
	logger.Error("error message", "id", 4)

	err = logger.Sync()
	require.NoError(t, err)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	logContent := string(content)

	// Assert that each message is present with the correct prefix
	assert.Contains(t, logContent, "[DEBUG] debug message")
	assert.Contains(t, logContent, `"id": 1`)
	assert.Contains(t, logContent, "[INFO] info message")
	assert.Contains(t, logContent, `"id": 2`)
	assert.Contains(t, logContent, "[WARN] warn message")
	assert.Contains(t, logContent, `"id": 3`)
	assert.Contains(t, logContent, "[ERROR] error message")
	assert.Contains(t, logContent, `"id": 4`)
}

// TestFatal is tricky because it calls os.Exit.
// We can't fully test the exit behavior without a complex setup (like exec).
// However, we can verify that the underlying zap method is called.
// For coverage, we'll just call it. A real-world test for this might use a mock or a different setup.
func TestFatal(t *testing.T) {
	// This test is limited. It ensures the Fatal method can be called without panic.
	// Since `zap.Fatal` calls `os.Exit`, we can't test the output in a normal test run.
	// The primary logic of `addLevelPrefix` is already covered in other tests.
	// We ensure our wrapper function `Fatal` is covered.
	assert.Panics(t, func() {
		// We need to override the default zapcore behavior for the test
		// This is a bit of a hack to prevent os.Exit from being called
		// and instead cause a panic that we can recover from.
		resetOnce()
		defer resetOnce()

		encoderConfig := zap.NewProductionEncoderConfig()
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(io.Discard), // write to nowhere
			zapcore.DebugLevel,
		)

		// Replace the fatal action with a panic
		hookedCore := zapcore.RegisterHooks(core, func(e zapcore.Entry) error {
			if e.Level == zapcore.FatalLevel {
				panic(fmt.Sprintf("fatal hook: %s", e.Message))
			}
			return nil
		})

		zapLogger := zap.New(hookedCore, zap.AddCaller(), zap.AddCallerSkip(1))
		logger := &Logger{zap: zapLogger.Sugar()}

		logger.Fatal("this is a fatal error")
	}, "Expected Fatal to cause a panic via the test hook")
}

// TestSync_NilLogger ensures that calling Sync on a nil logger does not panic.
func TestSync_NilLogger(t *testing.T) {
	// This tests the guard clause `if l.zap != nil`
	logger := &Logger{zap: nil}
	err := logger.Sync()
	assert.NoError(t, err, "Sync on a logger with a nil zap field should not return an error")
}

// TestAddLevelPrefix tests the internal helper function directly.
func TestAddLevelPrefix(t *testing.T) {
	msg := "test message"
	level := "level"
	expected := "[LEVEL] test message"
	result := addLevelPrefix(level, msg)
	assert.Equal(t, expected, result)
	assert.Equal(t, "[DEBUG] another", addLevelPrefix("debug", "another"))
}

// TestNewLogger_InstanceFailure is a theoretical test for the nil instance check.
// It's very hard to make the singleton's `once.Do` fail to set the instance.
// This test simulates the condition where `instance` remains nil after the `once.Do` block.
func TestNewLogger_InstanceFailure(t *testing.T) {
	// This is a white-box test for a condition that is unlikely to happen in practice.
	// We manually reset the instance to nil after `once.Do` would have run.
	once = sync.Once{} // Reset
	instance = nil     // Ensure it's nil

	// We can't directly inject a failure into once.Do, so we can't fully test the error path.
	// The `if instance == nil` check is a safeguard. The existing tests implicitly cover
	// the success path where it is not nil. We accept that this specific error return
	// is not directly testable without modifying the source code for testability.
	// However, if we could force instance to be nil:
	// _, err := NewLogger("info", "")
	// assert.Error(t, err)
	// assert.Equal(t, "failed to create logger instance", err.Error())
	// For now, we acknowledge this path exists but is untestable in the current code structure.
	t.Skip("Skipping test for untestable error condition without code modification.")
}
