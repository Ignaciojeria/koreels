package out

import "context"

// AudioUploader sube un WAV a almacenamiento y devuelve la URL pública.
type AudioUploader interface {
	UploadWAV(ctx context.Context, wavBytes []byte) (url string, err error)
}
