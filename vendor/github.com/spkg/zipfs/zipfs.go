// Package zipfs provides an implementation of the net/http.FileSystem
// interface based on the contents of a ZIP file. It also provides
// the FileServer function, which returns a net/http.Handler that
// serves static files from a ZIP file. This HTTP handler exploits
// the fact that most files are stored in a ZIP file using the
// deflate compression algorithm, and that most HTTP user agents will
// accept deflate as a content-encoding. When possible the HTTP
// handler will send the compressed file contents back to the
// user agent without having to decompress the ZIP file contents.
package zipfs
