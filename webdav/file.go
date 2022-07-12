// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"context"
	"path"
	"path/filepath"

	"github.com/gaoyb7/115drive-webdav/common/drive"
	"github.com/sirupsen/logrus"
)

type WalkFunc func(path string, info drive.File, err error) error

func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// Allowed values for depth are 0, 1 or infiniteDepth. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns filepath.SkipDir, walkFS will skip traversal of this node.
func walkFS(ctx context.Context, depth int, name string, fs drive.DriveClient, fi drive.File, walkFn WalkFunc) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(name, fi, nil)
	if err != nil {
		if fi.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !fi.IsDir() || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}

	// Read directory names.
	// f, err := fs.OpenFile(ctx, name, os.O_RDONLY, 0)
	// if err != nil {
	// 	return walkFn(name, info, err)
	// }
	// fileInfos, err := f.Readdir(0)
	// f.Close()
	// if err != nil {
	// 	return walkFn(name, info, err)
	// }
	files, err := fs.GetFiles(name)
	if err != nil {
		logrus.WithError(err).Errorf("call client.GetFiles fail, name: %s", name)
		return walkFn(name, fi, err)
	}

	for _, file := range files {
		filename := path.Join(name, file.GetName())
		if err != nil {
			logrus.WithError(err).Errorf("call client.GetFile fail, file_name: %s", filename)
			return err
			// if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
			// 	return err
			// }
		} else {
			err = walkFS(ctx, depth, filename, fs, file, walkFn)
			if err != nil {
				if !file.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}
