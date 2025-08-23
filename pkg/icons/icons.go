package icons

import (
	"fmt"
	"path"
	"strings"
)

type Icon int

const (
	iconUnknown Icon = iota
	Folder
	Archive
	Photo
	MusicNote
	DeriveFile
)

func (icon Icon) String() string {
	switch icon {
	case Folder:
		return "folder"
	case Archive:
		return "archive"
	case Photo:
		return "photo"
	case MusicNote:
		return "music_note"
	case DeriveFile:
		return "insert_drive_file"
	default:
		panic(fmt.Errorf("unknown icon: %d", icon))
	}
}

func Detect(fileName string) Icon {
	if strings.HasSuffix(fileName, "/") {
		return Folder
	}

	e := path.Ext(fileName)
	switch e {
	case ".tgz", ".gz", ".zip":
		return Archive
	case ".png", ".jpg", ".gif", ".svg":
		return Photo
	case ".mp3", ".wav":
		return MusicNote
	}

	return DeriveFile
}
