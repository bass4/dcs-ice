package unit

import (
	"testing"

	"github.com/yourusername/dcs-ice/pkg/models"
)

func TestFactCreation(t *testing.T) {
	fact := models.Fact{
		Event:      "unit_destroyed",
		Unit:       "convoy_alpha",
		Zone:       "BRAVO",
		AlertLevel: "red",
	}

	if fact.Event != "unit_destroyed" {
		t.Errorf("Expected event to be 'unit_destroyed', got '%s'", fact.Event)
	}

	if fact.Zone != "BRAVO" {
		t.Errorf("Expected zone to be 'BRAVO', got '%s'", fact.Zone)
	}
}
