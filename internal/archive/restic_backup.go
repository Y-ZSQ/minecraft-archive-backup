package archive

import (
	"minecraft-archive-backup/model/dto/database"
	"os/exec"
)

func ResticBackup(archive *database.Archive) <-chan *BackupMessage {
	cmd := NewResticCmd(exec.Command("restic", "backup", archive.Path,
		"--json",
		"--use-fs-snapshot",
		"-o", "vss.timeout=30s",
		//"--skip-if-unchanged",
	))

	return executeResticCommand(cmd)
}
