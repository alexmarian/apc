package handlers

import (
	"os"
	"testing"

	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Handlers log on their failure paths; main() normally sets this up.
	logging.Logger = zap.NewNop()
	os.Exit(m.Run())
}
