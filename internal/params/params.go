// Package params holds the pure transformations applied to SSM parameters:
// deriving environment-variable keys, merging parameters across services, and
// formatting them for output. None of this touches AWS.
package params

import (
	"strings"
	"time"
)

// Parameter is a single fetched SSM parameter value.
type Parameter struct {
	// Name is the full SSM name, e.g. "/staging/auth-service/one-key".
	Name    string
	Value   string
	Version int64
}

// Metadata is the subset of SSM parameter metadata shown by `describe`.
type Metadata struct {
	Name             string
	Version          int64
	LastModifiedUser string
	LastModifiedDate time.Time
}

// key returns the final "/"-delimited segment of an SSM name. For
// "/staging/auth-service/one-key" it returns "one-key".
func key(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// EnvKey converts an SSM parameter name into its environment-variable form:
// the final path segment, uppercased, with "-" replaced by "_".
// "/staging/svc/one-key" -> "ONE_KEY".
func EnvKey(name string) string {
	return strings.ReplaceAll(strings.ToUpper(key(name)), "-", "_")
}

// ShortName returns the final path segment of an SSM name, used as the display
// key in tables (e.g. "one-key").
func ShortName(name string) string {
	return key(name)
}

// Pair is a resolved environment-variable key/value.
type Pair struct {
	Key   string
	Value string
}

// Merge flattens parameters from multiple services into a single ordered list
// of env-var pairs. Services are processed left to right; the first service to
// define a given key wins, matching the documented conflict-resolution rule
// ("the natural ordering, left to right, is used to resolve key conflicts").
//
// perService is a slice-of-slices: perService[i] holds the parameters fetched
// for the i-th service, in the same order the services were given on the
// command line. Output order is deterministic: keys appear in the order first
// seen while scanning services left to right.
func Merge(perService [][]Parameter) []Pair {
	seen := make(map[string]struct{})
	var pairs []Pair

	for _, params := range perService {
		for _, p := range params {
			k := EnvKey(p.Name)
			if _, dup := seen[k]; dup {
				continue // earlier (left-most) service wins
			}
			seen[k] = struct{}{}
			pairs = append(pairs, Pair{Key: k, Value: p.Value})
		}
	}

	return pairs
}
