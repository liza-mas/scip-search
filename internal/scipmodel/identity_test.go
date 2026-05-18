package scipmodel

import "testing"

func TestParseIdentityDerivesPackageFieldsAndPreservesSymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		symbol     string
		descriptor string
		packageKey string
	}{
		{
			name:       "class descriptor",
			symbol:     "scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#",
			descriptor: "supervisor/Supervisor#",
			packageKey: "scip-go gomod github.com/liza-mas/liza .",
		},
		{
			name:       "method descriptor",
			symbol:     "scip-go gomod github.com/liza-mas/liza . supervisor/Run().",
			descriptor: "supervisor/Run().",
			packageKey: "scip-go gomod github.com/liza-mas/liza .",
		},
		{
			name:       "function descriptor",
			symbol:     "scip-go gomod github.com/liza-mas/scip-search . internal/query/Search().",
			descriptor: "internal/query/Search().",
			packageKey: "scip-go gomod github.com/liza-mas/scip-search .",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			identity, err := ParseIdentity(tt.symbol)
			if err != nil {
				t.Fatalf("ParseIdentity() error = %v", err)
			}

			if identity.Symbol != tt.symbol {
				t.Fatalf("Symbol = %q, want exact original %q", identity.Symbol, tt.symbol)
			}
			if identity.Scheme != "scip-go" {
				t.Fatalf("Scheme = %q, want %q", identity.Scheme, "scip-go")
			}
			if identity.PackageManager != "gomod" {
				t.Fatalf("PackageManager = %q, want %q", identity.PackageManager, "gomod")
			}
			if identity.PackageVersion != "." {
				t.Fatalf("PackageVersion = %q, want %q", identity.PackageVersion, ".")
			}
			if identity.Descriptor != tt.descriptor {
				t.Fatalf("Descriptor = %q, want %q", identity.Descriptor, tt.descriptor)
			}
			if identity.PackageKey() != tt.packageKey {
				t.Fatalf("PackageKey() = %q, want %q", identity.PackageKey(), tt.packageKey)
			}
		})
	}
}

func TestParseIdentityDerivesPackageName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		symbol      string
		packageName string
	}{
		{
			name:        "liza package",
			symbol:      "scip-go gomod github.com/liza-mas/liza . supervisor/Supervisor#",
			packageName: "github.com/liza-mas/liza",
		},
		{
			name:        "scip-search package",
			symbol:      "scip-go gomod github.com/liza-mas/scip-search . internal/query/Search().",
			packageName: "github.com/liza-mas/scip-search",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			identity, err := ParseIdentity(tt.symbol)
			if err != nil {
				t.Fatalf("ParseIdentity() error = %v", err)
			}

			if identity.PackageName != tt.packageName {
				t.Fatalf("PackageName = %q, want %q", identity.PackageName, tt.packageName)
			}
		})
	}
}

func TestIdentityMatchTextPrefersDisplayNameAndFallsBackToDescriptor(t *testing.T) {
	t.Parallel()

	identity, err := ParseIdentity("scip-go gomod github.com/liza-mas/liza . supervisor/Run().")
	if err != nil {
		t.Fatalf("ParseIdentity() error = %v", err)
	}

	if got := identity.MatchText("Run"); got != "Run" {
		t.Fatalf("MatchText(displayName) = %q, want %q", got, "Run")
	}
	if got := identity.MatchText(""); got != "supervisor/Run()." {
		t.Fatalf("MatchText(empty displayName) = %q, want descriptor fallback %q", got, "supervisor/Run().")
	}
}

func TestIsLocalSymbolIdentifiesSCIPLocalSymbols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		symbol string
		want   bool
	}{
		{
			name:   "local symbol",
			symbol: "local 0",
			want:   true,
		},
		{
			name:   "package bearing symbol",
			symbol: "scip-go gomod github.com/liza-mas/liza . supervisor/Run().",
			want:   false,
		},
		{
			name:   "local prefix without separator",
			symbol: "locality gomod example.com/pkg . pkg/Foo#",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsLocalSymbol(tt.symbol); got != tt.want {
				t.Fatalf("IsLocalSymbol(%q) = %v, want %v", tt.symbol, got, tt.want)
			}
		})
	}
}

func TestParseIdentityRejectsSymbolsWithoutDescriptor(t *testing.T) {
	t.Parallel()

	if _, err := ParseIdentity("scip-go gomod github.com/liza-mas/liza ."); err == nil {
		t.Fatal("ParseIdentity() error = nil, want error")
	}
}
