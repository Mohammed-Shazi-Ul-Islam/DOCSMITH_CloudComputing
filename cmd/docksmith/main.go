package main

import (
	"fmt"
	"os"
	"strings"

	"docksmith/internal/runtime"
	"docksmith/internal/state"
	"docksmith/internal/builder"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	var exitCode int

	switch os.Args[1] {
	case "build":
		exitCode, err = cmdBuild(os.Args[2:])
	case "images":
		exitCode, err = cmdImages()
	case "rmi":
		exitCode, err = cmdRmi(os.Args[2:])
	case "run":
		exitCode, err = cmdRun(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func printUsage() {
	fmt.Print(`Usage: docksmith <command> [options]

Commands:
  build -t <name:tag> <context_dir> [--no-cache]   Build an image
  images                                             List local images
  rmi <name:tag>                                     Remove an image
  run [-e KEY=VALUE] <name:tag> [command...]         Run a container
`)
}

func cmdBuild(args []string) (int, error) {
	var tag, contextDir string
	noCache := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-t", "--tag":
			i++
			if i >= len(args) {
				return 1, fmt.Errorf("[CLI ERROR] -t requires an argument")
			}
			tag = args[i]
		case "--no-cache":
			noCache = true
		default:
			if contextDir == "" {
				contextDir = args[i]
			}
		}
	}

	if tag == "" {
		return 1, fmt.Errorf("[CLI ERROR] -t <name:tag> is required")
	}
	if contextDir == "" {
		return 1, fmt.Errorf("[CLI ERROR] context directory is required")
	}

	name, imageTag, err := parseImageRef(tag)
	if err != nil {
		return 1, err
	}

	_, err = builder.BuildImage(contextDir, name, imageTag, noCache)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func cmdImages() (int, error) {
	manifests, err := state.ListManifests()
	if err != nil {
		return 1, err
	}
	if len(manifests) == 0 {
		fmt.Println("No images found.")
		return 0, nil
	}
	fmt.Printf("%-20s %-12s %-14s %s\n", "NAME", "TAG", "ID", "CREATED")
	for _, m := range manifests {
		id := m.Digest
		if len(id) > 12 {
			id = id[:12]
		}
		fmt.Printf("%-20s %-12s %-14s %s\n", m.Name, m.Tag, id, m.Created)
	}
	return 0, nil
}

func cmdRmi(args []string) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("[CLI ERROR] rmi requires an image reference")
	}
	name, tag, err := parseImageRef(args[0])
	if err != nil {
		return 1, err
	}
	removed, err := state.RemoveImage(name, tag)
	if err != nil {
		return 1, err
	}
	if !removed {
		fmt.Printf("Image not found: %s:%s\n", name, tag)
		return 1, nil
	}
	fmt.Printf("Removed image %s:%s\n", name, tag)
	return 0, nil
}

func cmdRun(args []string) (int, error) {
	var envOverrides []string
	var imageRef string
	var commandOverride []string

	for i := 0; i < len(args); i++ {
		if args[i] == "-e" || args[i] == "--env" {
			i++
			if i >= len(args) {
				return 1, fmt.Errorf("[CLI ERROR] -e requires KEY=VALUE argument")
			}
			envOverrides = append(envOverrides, args[i])
		} else if strings.HasPrefix(args[i], "-e=") {
			envOverrides = append(envOverrides, strings.TrimPrefix(args[i], "-e="))
		} else if imageRef == "" {
			imageRef = args[i]
		} else {
			commandOverride = append(commandOverride, args[i])
		}
	}

	if imageRef == "" {
		return 1, fmt.Errorf("[CLI ERROR] run requires an image reference")
	}

	name, tag, err := parseImageRef(imageRef)
	if err != nil {
		return 1, err
	}

	manifest, err := state.LoadManifest(name, tag)
	if err != nil {
		return 1, err
	}
	if manifest == nil {
		return 1, fmt.Errorf("[RUNTIME ERROR] image not found: %s:%s", name, tag)
	}

	envMap := make(map[string]string)
	for _, pair := range envOverrides {
		if idx := strings.Index(pair, "="); idx >= 0 {
			envMap[pair[:idx]] = pair[idx+1:]
		} else {
			return 1, fmt.Errorf("[CLI ERROR] -e must be KEY=VALUE, got: %q", pair)
		}
	}

	code, err := runtime.RunImage(manifest, commandOverride, envMap)
	if err != nil {
		return 1, err
	}
	fmt.Printf("Container exited with code %d\n", code)
	return code, nil
}

func parseImageRef(ref string) (name, tag string, err error) {
	if idx := strings.Index(ref, ":"); idx >= 0 {
		name = strings.TrimSpace(ref[:idx])
		tag = strings.TrimSpace(ref[idx+1:])
	} else {
		name = strings.TrimSpace(ref)
		tag = "latest"
	}
	if name == "" || tag == "" {
		return "", "", fmt.Errorf("[CLI ERROR] invalid image reference: %q", ref)
	}
	return name, tag, nil
}
