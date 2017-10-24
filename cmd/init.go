// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cmd provides supporting functions for the matcha command line tool.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Init(flags *Flags) error {
	start := time.Now()

	// Parse targets
	targets := ParseTargets(flags.BuildTargets)

	// Get $GOPATH/pkg/matcha directory
	matchaPkgPath, err := MatchaPkgPath(flags)
	if err != nil {
		return err
	}

	// Delete $GOPATH/pkg/matcha
	if err := RemoveAll(flags, matchaPkgPath); err != nil {
		return err
	}

	// Make $GOPATH/pkg/matcha
	if err := Mkdir(flags, matchaPkgPath); err != nil {
		return err
	}

	// Make $WORK
	tmpdir, err := NewTmpDir(flags, "")
	if err != nil {
		return err
	}
	defer RemoveAll(flags, tmpdir)

	// Begin iOS
	if _, ok := targets["ios"]; ok {
		// Validate Xcode installation
		if err := validateXcodeInstall(flags); err != nil {
			return err
		}

		// Install standard libraries for cross compilers.
		var env []string

		if _, ok := targets["ios/arm"]; ok {
			if env, err = DarwinArmEnv(flags); err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}

		if _, ok := targets["ios/arm64"]; ok {
			if env, err = DarwinArm64Env(flags); err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}

		if _, ok := targets["ios/386"]; ok {
			if env, err = Darwin386Env(flags); err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env, "-tags=ios"); err != nil {
				return err
			}
		}

		if _, ok := targets["ios/amd64"]; ok {
			if env, err = DarwinAmd64Env(flags); err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env, "-tags=ios"); err != nil {
				return err
			}
		}
	}

	// Begin android
	if _, ok := targets["android"]; ok {
		// Validate Android installation
		if err := validateAndroidInstall(flags); err != nil {
			return err
		}

		// Install standard libraries for cross compilers.
		if _, ok := targets["android/arm"]; ok {
			env, err := androidEnv(flags, "arm")
			if err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}

		if _, ok := targets["android/arm64"]; ok {
			env, err := androidEnv(flags, "arm64")
			if err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}

		if _, ok := targets["android/386"]; ok {
			env, err := androidEnv(flags, "386")
			if err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}

		if _, ok := targets["android/amd64"]; ok {
			env, err := androidEnv(flags, "amd64")
			if err != nil {
				return err
			}
			if err := InstallPkg(flags, matchaPkgPath, tmpdir, "std", env); err != nil {
				return err
			}
		}
	}

	// Write Go Version to $GOPATH/pkg/matcha/version
	goversion, err := GoVersion(flags)
	if err != nil {
		return nil
	}
	verpath := filepath.Join(matchaPkgPath, "version")
	if err := WriteFile(flags, verpath, bytes.NewReader(goversion)); err != nil {
		return err
	}

	// Timing
	if flags.BuildV {
		took := time.Since(start) / time.Second * time.Second
		fmt.Fprintf(os.Stderr, "Build took %s.\n", took)
	}
	fmt.Fprintf(os.Stderr, "Matcha initialized.\n")
	return nil
}

// Build package with properties.
func InstallPkg(f *Flags, matchaPkgPath, temp string, pkg string, env []string, args ...string) error {
	pkgPath, err := PkgPath(f, matchaPkgPath, env)
	if err != nil {
		return err
	}

	tOS, tArch := FindEnv(env, "GOOS"), FindEnv(env, "GOARCH")
	if tOS != "" && tArch != "" {
		if f.BuildV {
			fmt.Fprintf(os.Stderr, "\n# Installing %s for %s/%s.\n", pkg, tOS, tArch)
		}
		args = append(args, "-pkgdir="+pkgPath)
	} else {
		if f.BuildV {
			fmt.Fprintf(os.Stderr, "\n# Installing %s.\n", pkg)
		}
	}

	cmd := exec.Command("go", "install")
	cmd.Args = append(cmd.Args, args...)
	if f.BuildV {
		cmd.Args = append(cmd.Args, "-v")
	}
	if f.BuildX {
		cmd.Args = append(cmd.Args, "-x")
	}
	if f.BuildWork {
		cmd.Args = append(cmd.Args, "-work")
	}
	cmd.Args = append(cmd.Args, pkg)
	cmd.Env = append([]string{}, env...)
	return RunCmd(f, temp, cmd)
}
