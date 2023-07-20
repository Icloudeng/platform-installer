package utilities

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"smatflow/platform-installer/pkg/files"
	"smatflow/platform-installer/pkg/structs"
)

func SendNotification(notifier structs.Notifier) {

	// Clear up
	cmd := exec.Command(
		"bash", "notifier.sh",
		"--status", notifier.Status,
		"--details", notifier.Details,
		"--logs", base64.StdEncoding.EncodeToString([]byte(notifier.Logs)),
		"--metadata", base64.StdEncoding.EncodeToString([]byte(notifier.Metadata)),
	)

	cmd.Dir = files.ProvisionerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	fmt.Println("Notification Send !")
}