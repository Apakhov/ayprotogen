package parse

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	modulePrefix          = []byte("\nmodule ")
	pkgPathFromGoModCache = make(map[string]string)
)

var goModPathCache = struct {
	paths map[string]string
	sync.RWMutex
}{
	paths: make(map[string]string),
}

func isGoMod(pkgdir string) (bool, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Dir = pkgdir

	stdout, err := cmd.Output()
	if err != nil {
		return false, err
	}

	goModPath := string(bytes.TrimSpace(stdout))
	return strings.Contains(goModPath, "go.mod"), nil
}

func goModPath(fname string, isDir bool) (string, error) {
	root := fname
	if !isDir {
		root = filepath.Dir(fname)
	}

	goModPathCache.RLock()
	goModPath, ok := goModPathCache.paths[root]
	goModPathCache.RUnlock()
	if ok {
		return goModPath, nil
	}

	defer func() {
		goModPathCache.Lock()
		goModPathCache.paths[root] = goModPath
		goModPathCache.Unlock()
	}()

	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Dir = root

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	goModPath = string(bytes.TrimSpace(stdout))

	return goModPath, nil
}

func getPkgPathFromGoMod(fname string, isDir bool, goModPath string) (string, error) {
	modulePath := getModulePath(goModPath)
	if modulePath == "" {
		return "", fmt.Errorf("cannot determine module path from %s", goModPath)
	}

	rel := path.Join(modulePath, filepath.ToSlash(strings.TrimPrefix(fname, filepath.Dir(goModPath))))

	if !isDir {
		return path.Dir(rel), nil
	}

	return path.Clean(rel), nil
}

func getModulePath(goModPath string) string {
	pkgPath, ok := pkgPathFromGoModCache[goModPath]
	if ok {
		return pkgPath
	}

	defer func() {
		pkgPathFromGoModCache[goModPath] = pkgPath
	}()

	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	var i int
	if bytes.HasPrefix(data, modulePrefix[1:]) {
		i = 0
	} else {
		i = bytes.Index(data, modulePrefix)
		if i < 0 {
			return ""
		}
		i++
	}
	line := data[i:]

	// Cut line at \n, drop trailing \r if present.
	if j := bytes.IndexByte(line, '\n'); j >= 0 {
		line = line[:j]
	}
	if line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	line = line[len("module "):]

	// If quoted, unquote.
	pkgPath = strings.TrimSpace(string(line))
	if pkgPath != "" && pkgPath[0] == '"' {
		s, err := strconv.Unquote(pkgPath)
		if err != nil {
			return ""
		}
		pkgPath = s
	}
	return pkgPath
}
