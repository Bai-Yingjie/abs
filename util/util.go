package util

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/abs-lang/abs/object"
)

// Checks whether the element e is in the
// list of strings s
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)

	return err == nil
}

// ExpandPath (path) resolves leading "~/" to user's HomeDir
// returns expanded path, error
func ExpandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

// GetEnvVar (varName, defaultVal)
// Return the varName value from the ABS env, or OS env, or default value in that order
func GetEnvVar(env *object.Environment, varName, defaultVal string) string {
	var ok bool
	var value string
	valueObj, ok := env.Get(varName)
	if ok {
		value = valueObj.Inspect()
	} else {
		value = os.Getenv(varName)
		if len(value) == 0 {
			value = defaultVal
		}
	}
	return value
}

// InterpolateStringVars (str, env)
// return input string with $vars interpolated from environment
func InterpolateStringVars(str string, env *object.Environment) string {
	// Match all strings preceded by
	// a $ or a \$
	re := regexp.MustCompile("(\\\\)?\\$(\\{)?([a-zA-Z_0-9]{1,})(\\})?")
	str = re.ReplaceAllStringFunc(str, func(m string) string {
		// If the string starts with a backslash,
		// that's an escape, so we should replace
		// it with the remaining portion of the match.
		// \$VAR becomes $VAR
		if string(m[0]) == "\\" {
			return m[1:]
		}

		// If the string starts with $, then
		// it's an interpolation. Let's
		// replace $VAR with the variable
		// named VAR in the ABS' environment.
		// We need to support both ${var}
		// and $var.
		varName := ""
		if m[1] == '{' {
			// If you type a variable wrong, forgetting the
			// closing bracket, we simply return it to you:
			// eg "my ${variable"
			if m[len(m)-1] != '}' {
				return m
			}

			varName = m[2 : len(m)-1]
		} else {
			varName = m[1:]
		}

		v, ok := env.Get(varName)

		// If the variable is not found, we
		// just dump an empty string
		if !ok {
			return ""
		}
		return v.Inspect()
	})

	return str
}

// UniqueStrings takes an input list of strings
// and returns a version without duplicate values
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UnaliasPath translates a path alias
// to the full path in the filesystem.
func UnaliasPath(path string, packageAlias map[string]string) string {
	// An alias can come in different forms:
	//  - package
	//  - package/file.abs
	// but we only really need to resolve the
	// first path in the alias.
	parts := strings.Split(path, string(os.PathSeparator))

	if len(parts) < 1 {
		return path
	}

	if packageAlias[parts[0]] != "" {
		// If we are able to resolve a path, then
		// we should join in back with the rest of the
		// paths
		p := []string{packageAlias[parts[0]]}
		p = append(p, parts[1:]...)
		path = filepath.Join(p...)
	}
	return appendIndexFile(path)
}

// If our path didn't end with an ABS file (.abs),
// let's assume it's a directory and we will
// auto-include the index.abs file from it
func appendIndexFile(path string) string {
	if filepath.Ext(path) != ".abs" {
		return filepath.Join(path, "index.abs")
	}

	return path
}

// Mapify converts a list of objects to a map.
// This is useful when you want to test whether
// elements of a list are present in another list:
// You can mapify the first one and check whether
// elements of the second one would occupy the same
// key in the map.
func Mapify(list []object.Object) map[string]object.Object {
	m := make(map[string]object.Object)

	for _, v := range list {
		m[object.GenerateEqualityString(v)] = v
	}

	return m
}
