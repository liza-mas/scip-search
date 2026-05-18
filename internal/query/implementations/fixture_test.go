package implementations_test

import (
	"testing"

	"scip-search/internal/traversal/traversaltest"
)

const implementationAbsentSymbol = "scip-go gomod example.com/fixture . missing/Absent#"

func loadImplementationFixture(t testing.TB) traversaltest.Fixture {
	t.Helper()

	return traversaltest.LoadSharedFixture(t)
}
