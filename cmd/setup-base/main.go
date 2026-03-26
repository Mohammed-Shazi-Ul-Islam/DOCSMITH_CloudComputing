package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docksmith/internal/layers"
	"docksmith/internal/models"
	"docksmith/internal/paths"
	"docksmith/internal/state"
)

func main() {
	image := "alpine:latest"
	var rootfsDir, rootfsTar string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--image":
			i++
			image = args[i]
		case "--rootfs-dir":
			i++
			rootfsDir = args[i]
		case "--rootfs-tar":
			i++
			rootfsTar = args[i]
		}
	}

	if (rootfsDir == "") == (rootfsTar == "") {
		fmt.Fprintln(os.Stderr, "Provide exactly one of --rootfs-dir or --rootfs-tar")
		os.Exit(1)
	}

	if err := paths.EnsureStateDirs(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var dir string
	if rootfsDir != "" {
		dir = rootfsDir
	} else {
		var err error
		dir, err = extractTarToTemp(rootfsTar)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer os.RemoveAll(dir)
	}

	if err := importRootfsDir(image, dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func importRootfsDir(imageName, rootfsDir string) error {
	allPaths, err := layers.CollectAllPaths(rootfsDir)
	if err != nil {
		return err
	}
	tarBytes, err := layers.CreateDeltaTar(rootfsDir, allPaths)
	if err != nil {
		return err
	}
	digest, err := layers.StoreLayer(tarBytes)
	if err != nil {
		return err
	}

	name, tag := splitImageRef(imageName)
	manifest := models.ImageManifest{
		Name:    name,
		Tag:     tag,
		Digest:  "",
		Created: time.Now().UTC().Format(time.RFC3339Nano),
		Config: models.ImageConfig{
			Env:        []string{},
			Cmd:        []string{"/bin/sh"},
			WorkingDir: "/",
		},
		Layers: []models.LayerEntry{
			{
				Digest:    digest,
				Size:      int64(len(tarBytes)),
				CreatedBy: "base layer import",
			},
		},
	}

	saved, err := state.SaveManifest(manifest)
	if err != nil {
		return err
	}
	short := saved.Digest
	if len(short) > 19 {
		short = short[:19]
	}
	fmt.Printf("Imported base image %s:%s (%s)\n", saved.Name, saved.Tag, short)
	return nil
}

func extractTarToTemp(tarPath string) (string, error) {
	tmp, err := os.MkdirTemp("", "docksmith_base_")
	if err != nil {
		return "", err
	}

	f, err := os.Open(tarPath)
	if err != nil {
		os.RemoveAll(tmp)
		return "", err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(tmp)
			return "", err
		}
		target := filepath.Join(tmp, filepath.FromSlash(hdr.Name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, os.FileMode(hdr.Mode)|0755)
		case tar.TypeSymlink:
			os.MkdirAll(filepath.Dir(target), 0755)
			os.Symlink(hdr.Linkname, target)
		default:
			os.MkdirAll(filepath.Dir(target), 0755)
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				continue
			}
			io.Copy(out, tr)
			out.Close()
		}
	}
	return tmp, nil
}

func splitImageRef(ref string) (name, tag string) {
	if idx := strings.Index(ref, ":"); idx >= 0 {
		return ref[:idx], ref[idx+1:]
	}
	return ref, "latest"
}
