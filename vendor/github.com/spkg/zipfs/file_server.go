package zipfs

// Some of the functions in this file are adapted from private
// functions in the standard library net/http package.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

import (
	"archive/zip"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// FileServer returns a HTTP handler that serves
// HTTP requests with the contents of the ZIP file system.
// It provides slightly better performance than the
// http.FileServer implementation because it serves compressed content
// to clients that can accept the "deflate" compression algorithm.
func FileServer(fs *FileSystem) http.Handler {
	h := &fileHandler{
		fs: fs,
	}

	return h
}

type fileHandler struct {
	fs *FileSystem
}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	serveFile(w, r, h.fs, path.Clean(upath), true)
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs *FileSystem, name string, redirect bool) {
	const indexPage = "/index.html"

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")
		return
	}

	d, err := fs.openFileInfo(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	if redirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return
			}
		}
	}

	// use contents of index.html for directory, if present
	if d.IsDir() {
		index := strings.TrimSuffix(name, "/") + indexPage
		dd, err := fs.openFileInfo(index)
		if err == nil {
			d = dd
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		// Unlike the standard library implementation, directory
		// listing is prohibited.
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// serveContent will check modification time and ETag
	serveContent(w, r, fs, d)
}

func serveContent(w http.ResponseWriter, r *http.Request, fs *FileSystem, fi *fileInfo) {
	if checkLastModified(w, r, fi.ModTime()) {
		return
	}

	// Set the Etag header in the response before calling checkETag.
	// The checkETag function obtains the files ETag from the response header.
	w.Header().Set("Etag", calcEtag(fi.zipFile))
	rangeReq, done := checkETag(w, r, fi.ModTime())
	if done {
		return
	}
	if rangeReq != "" {
		// Range request requires seeking, so at this point create a temporary
		// file and let the standard library serve it.
		f := fi.openReader(r.URL.Path)
		defer f.Close()
		f.createTempFile()
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f.file)
		return
	}

	setContentType(w, fi.Name())

	switch fi.zipFile.Method {
	case zip.Store:
		serveIdentity(w, r, fi.zipFile)
	case zip.Deflate:
		serveDeflate(w, r, fi.zipFile, fs.readerAt)
	default:
		http.Error(w, fmt.Sprintf("unsupported zip method: %d", fi.zipFile.Method), http.StatusInternalServerError)
	}
}

// serveIdentity serves a zip file in identity content encoding .
func serveIdentity(w http.ResponseWriter, r *http.Request, zf *zip.File) {
	// TODO: need to check if the client explicitly refuses to accept
	// identity encoding (Accept-Encoding: identity;q=0), but this is
	// going to be very rare.

	reader, err := zf.Open()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer reader.Close()

	size := zf.FileInfo().Size()
	w.Header().Del("Content-Encoding")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	if r.Method != "HEAD" {
		io.CopyN(w, reader, int64(size))
	}
}

// serveDeflat serves a zip file in deflate content-encoding if the
// user agent can accept it. Otherwise it calls serveIdentity.
func serveDeflate(w http.ResponseWriter, r *http.Request, f *zip.File, readerAt io.ReaderAt) {
	acceptEncoding := r.Header.Get("Accept-Encoding")

	// TODO: need to parse the accept header to work out if the
	// client is explicitly forbidding deflate (ie deflate;q=0)
	acceptsDeflate := strings.Contains(acceptEncoding, "deflate")
	if !acceptsDeflate {
		// client will not accept deflate, so serve as identity
		serveIdentity(w, r, f)
		return
	}

	contentLength := int64(f.CompressedSize64)
	if contentLength == 0 {
		contentLength = int64(f.CompressedSize)
	}
	w.Header().Set("Content-Encoding", "deflate")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	if r.Method == "HEAD" {
		return
	}

	var written int64
	remaining := contentLength
	offset, err := f.DataOffset()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// re-use buffers to reduce stress on GC
	buf := bufPool.Get()
	defer bufPool.Free(buf)

	// loop to write the raw deflated content to the client
	for remaining > 0 {
		size := len(buf)
		if int64(size) > remaining {
			size = int(remaining)
		}

		b := buf[:size]
		_, err := readerAt.ReadAt(b, offset)
		if err != nil {
			if written == 0 {
				// have not written anything to the client yet, so we can send an error
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
			}
			return
		}
		if _, err := w.Write(b); err != nil {
			// Cannot write an error to the client because, er,  we just
			// failed to write to the client.
			return
		}
		written += int64(size)
		remaining -= int64(size)
		offset += int64(size)
	}
}

func setContentType(w http.ResponseWriter, filename string) {
	ctypes, haveType := w.Header()["Content-Type"]
	var ctype string
	if !haveType {
		ctype = mime.TypeByExtension(filepath.Ext(path.Base(filename)))
		if ctype == "" {
			// the standard library sniffs content to decide whether it is
			// binary or text, but this requires a ReaderSeeker, and we
			// only have a reader from the zip file. Assume binary.
			ctype = "application/octet-stream"
		}
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}
	if ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}
}

// calcEtag calculates an ETag value for a given zip file based on
// the file's CRC and its length.
func calcEtag(f *zip.File) string {
	size := f.UncompressedSize64
	if size == 0 {
		size = uint64(f.UncompressedSize)
	}
	etag := uint64(f.CRC32) ^ (uint64(size&0xffffffff) << 32)

	// etag should always be in double quotes
	return fmt.Sprintf(`"%x"`, etag)
}

var unixEpochTime = time.Unix(0, 0)

// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() || modtime.Equal(unixEpochTime) {
		// If the file doesn't have a modtime (IsZero), or the modtime
		// is obviously garbage (Unix time == 0), then ignore modtimes
		// and don't process the If-Modified-Since header.
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

// checkETag implements If-None-Match and If-Range checks.
//
// The ETag or modtime must have been previously set in the
// ResponseWriter's headers.  The modtime is only compared at second
// granularity and may be the zero value to mean unknown.
//
// The return value is the effective request "Range" header to use and
// whether this request is now considered done.
func checkETag(w http.ResponseWriter, r *http.Request, modtime time.Time) (rangeReq string, done bool) {
	etag := w.Header().Get("Etag")
	rangeReq = r.Header.Get("Range")

	// Invalidate the range request if the entity doesn't match the one
	// the client was expecting.
	// "If-Range: version" means "ignore the Range: header unless version matches the
	// current file."
	// We only support ETag versions.
	// The caller must have set the ETag on the response already.
	if ir := r.Header.Get("If-Range"); ir != "" && ir != etag {
		// The If-Range value is typically the ETag value, but it may also be
		// the modtime date. See golang.org/issue/8367.
		timeMatches := false
		if !modtime.IsZero() {
			if t, err := http.ParseTime(ir); err == nil && t.Unix() == modtime.Unix() {
				timeMatches = true
			}
		}
		if !timeMatches {
			rangeReq = ""
		}
	}

	if inm := r.Header.Get("If-None-Match"); inm != "" {
		// Must know ETag.
		if etag == "" {
			return rangeReq, false
		}

		// TODO(bradfitz): non-GET/HEAD requests require more work:
		// sending a different status code on matches, and
		// also can't use weak cache validators (those with a "W/
		// prefix).  But most users of ServeContent will be using
		// it on GET or HEAD, so only support those for now.
		if r.Method != "GET" && r.Method != "HEAD" {
			return rangeReq, false
		}

		// TODO(bradfitz): deal with comma-separated or multiple-valued
		// list of If-None-match values.  For now just handle the common
		// case of a single item.
		if inm == etag || inm == "*" {
			h := w.Header()
			delete(h, "Content-Type")
			delete(h, "Content-Length")
			w.WriteHeader(http.StatusNotModified)
			return "", true
		}
	}
	return rangeReq, false
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if pathErr, ok := err.(*os.PathError); ok {
		err = pathErr.Err
	}
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
