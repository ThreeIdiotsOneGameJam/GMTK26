//go:build web

package render

// Animated edge shaders require per-vertex normals, which cannot use the web
// hex batch. At distant zoom levels the edges are only a few pixels wide, so
// omit them before their cross-WASM call count can dominate the frame.
const maxDetailedEdgeTiles = 1200
