package unpack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/candrewlee14/webman/utils"
	"github.com/mholt/archives"
)

func Unpack(src string, pkg string, stem string, hasRoot bool) error {
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	err := os.MkdirAll(pkgDir, 0o755)
	if err != nil {
		return fmt.Errorf("unable to create dir %q: %v", pkgDir, err)
	}
	pkgDest := filepath.Join(pkgDir, stem)
	if hasRoot {
		tmpPkgDir := filepath.Join(utils.WebmanTmpDir, pkg)
		if err := os.MkdirAll(tmpPkgDir, 0o755); err != nil {
			return fmt.Errorf("unable to create dir %q: %v", tmpPkgDir, err)
		}
		// Call the refactored unpack function
		if err := unpack(pkg, src, tmpPkgDir); err != nil {
			// Attempt to clean up temporary directory on error
			os.RemoveAll(tmpPkgDir)
			return fmt.Errorf("failed to extract file: %v", err)
		}
		// Read the contents of the temporary directory
		entries, err := os.ReadDir(tmpPkgDir)
		if err != nil {
			os.RemoveAll(tmpPkgDir)
			return fmt.Errorf("unable to read dir %q: %v", tmpPkgDir, err)
		}
		// Ensure there's exactly one entry (the extracted folder)
		if len(entries) != 1 {
			os.RemoveAll(tmpPkgDir)
			return fmt.Errorf("expected unzipped archive to have a single root folder, found %d entries", len(entries))
		}
		extractFolder := filepath.Join(tmpPkgDir, entries[0].Name())
		// Move the extracted folder to the final destination
		if err = os.Rename(extractFolder, pkgDest); err != nil {
			os.RemoveAll(tmpPkgDir) // Clean up if rename fails
			return fmt.Errorf("unable to move %q to %q: %v", extractFolder, pkgDest, err)
		}
		// Successfully moved, now remove the (now empty) temporary parent directory
		if err := os.Remove(tmpPkgDir); err != nil {
			// Log or handle minor cleanup error if necessary, but don't fail the unpack
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary directory %s: %v\n", tmpPkgDir, err)
		}
	} else {
		if err := os.MkdirAll(pkgDest, 0o777); err != nil {
			return fmt.Errorf("unable to create pkg destination dir %q: %v", pkgDest, err)
		}
		if err := unpack(pkg, src, pkgDest); err != nil {
			return fmt.Errorf("failed to extract file: %v", err)
		}
	}
	return nil
}

func unpack(pkg, src, dest string) error {
	ctx := context.Background()

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer file.Close()

	format, stream, err := archives.Identify(ctx, src, file)
	if err != nil {
		return fmt.Errorf("failed to identify archive format for %s: %w", src, err)
	}
	if stream == nil {
		// If Identify returns a nil stream, we might need to re-open the file for the extractor/decompressor
		// or it means the format itself doesn't use the stream from Identify directly.
		// For safety, let's re-open. Some format handlers might manage their own stream.
		file.Close() // Close the one used for Identify
		file, err = os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to re-open source file %s for extraction: %w", src, err)
		}
		defer file.Close() // Defer close for this new file instance
		stream = file // Use this new file instance as the stream
	}


	switch f := format.(type) {
	case archives.Extractor:
		handler := func(ctx context.Context, fileInfo archives.FileInfo) error {
			destPath := filepath.Join(dest, fileInfo.NameInArchive)
			if fileInfo.IsDir() {
				// Ensure parent directory exists
				if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
					return fmt.Errorf("failed to create parent directory for %s: %w", destPath, err)
				}
				// Create the directory itself
				if err := os.MkdirAll(destPath, fileInfo.Mode()); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", destPath, err)
				}
				return nil
			}

			// Ensure parent directory exists for files too
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for file %s: %w", destPath, err)
			}

			fileInArchive, err := fileInfo.Open()
			if err != nil {
				return fmt.Errorf("failed to open file %s in archive: %w", fileInfo.NameInArchive, err)
			}
			defer fileInArchive.Close()

			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, fileInArchive); err != nil {
				return fmt.Errorf("failed to copy content to %s: %w", destPath, err)
			}

			// Close the output file before changing permissions
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("failed to close destination file %s: %w", destPath, err)
			}
			// fileInArchive is closed by its defer

			if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
				return fmt.Errorf("failed to set permissions for %s: %w", destPath, err)
			}
			return nil
		}
		if err := f.Extract(ctx, stream, handler); err != nil {
			return fmt.Errorf("failed to extract archive %s: %w", src, err)
		}

	case archives.Decompressor:
		var finalFileName string
		if utils.GOOS == "windows" {
			finalFileName = pkg + ".exe"
		} else {
			finalFileName = pkg
		}
		fullDestPath := filepath.Join(dest, finalFileName)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fullDestPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", fullDestPath, err)
		}

		outFile, err := os.Create(fullDestPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", fullDestPath, err)
		}
		defer outFile.Close()

		decompressedReader, err := f.OpenReader(stream)
		if err != nil {
			return fmt.Errorf("failed to open decompressor for %s: %w", src, err)
		}
		defer decompressedReader.Close()

		if _, err := io.Copy(outFile, decompressedReader); err != nil {
			return fmt.Errorf("failed to decompress content to %s: %w", fullDestPath, err)
		}

		// Close file before chmod
		if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file %s before chmod: %w", fullDestPath, err)
	}


		if err := os.Chmod(fullDestPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions for %s: %w", fullDestPath, err)
		}

	default:
		return fmt.Errorf("unsupported archive format for %s: %T", src, format)
	}

	return nil
}
