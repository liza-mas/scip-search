package aggregate

import (
	"fmt"
	"strings"
)

func ParseArgs(args []string) (Options, error) {
	var options Options
	var pendingRoot string
	expectIndex := false

	for position := 0; position < len(args); position++ {
		arg := args[position]
		switch arg {
		case "--project-root":
			value, ok := flagValue(args, &position)
			if !ok {
				return Options{}, NewValidationError("--project-root requires a value")
			}
			if options.ProjectRoot != "" {
				return Options{}, NewValidationError("--project-root can only be provided once")
			}
			options.ProjectRoot = value
		case "--out":
			value, ok := flagValue(args, &position)
			if !ok {
				return Options{}, NewValidationError("--out requires a value")
			}
			if options.OutPath != "" {
				return Options{}, NewValidationError("--out can only be provided once")
			}
			options.OutPath = value
		case "--root":
			if expectIndex {
				return Options{}, NewValidationError("--root must be followed by --index before another --root")
			}
			value, ok := flagValue(args, &position)
			if !ok {
				return Options{}, NewValidationError("--root requires a value")
			}
			pendingRoot = value
			expectIndex = true
		case "--index":
			if !expectIndex {
				return Options{}, NewValidationError("--index must follow --root")
			}
			value, ok := flagValue(args, &position)
			if !ok {
				return Options{}, NewValidationError("--index requires a value")
			}
			options.Pairs = append(options.Pairs, Pair{Root: pendingRoot, IndexPath: value})
			pendingRoot = ""
			expectIndex = false
		default:
			return Options{}, NewValidationError(fmt.Sprintf("aggregate-index only accepts --project-root, --root, --index, and --out; got %s", arg))
		}
	}
	if expectIndex {
		return Options{}, NewValidationError("--root must be followed by --index")
	}
	if strings.TrimSpace(options.ProjectRoot) == "" {
		return Options{}, NewValidationError("missing --project-root")
	}
	if strings.TrimSpace(options.OutPath) == "" {
		return Options{}, NewValidationError("missing --out")
	}
	if len(options.Pairs) < 1 {
		return Options{}, NewValidationError("aggregate-index requires at least one --root/--index input pair")
	}
	if _, _, err := normalizeProjectRoot(options.ProjectRoot); err != nil {
		return Options{}, err
	}
	for _, pair := range options.Pairs {
		if _, err := cleanSourceRoot(pair.Root); err != nil {
			return Options{}, err
		}
	}
	return options, nil
}

func flagValue(args []string, position *int) (string, bool) {
	if *position+1 >= len(args) || args[*position+1] == "" || strings.HasPrefix(args[*position+1], "--") {
		return "", false
	}
	*position = *position + 1
	return args[*position], true
}
