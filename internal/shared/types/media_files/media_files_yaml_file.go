// Package media_files defines YAML representations of media file collections.
package media_files

import "github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"

type MediaFilesYamlFile struct {
	Files []media_file.MediaFileData
}
