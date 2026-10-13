//go:build init

package assets

// GetAsset returns the bundled bytes for an asset filename
// (e.g. "Ayush_Anand.jpg"). Used by init-time seeding that
// uploads bundled photos to the object store.
func GetAsset(name string) ([]byte, bool) {
	data, ok := assets[name]
	return data, ok
}
