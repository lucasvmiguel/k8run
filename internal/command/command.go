package command

import "fmt"

func pvcName(name string) string {
	return fmt.Sprintf("%s-app-pvc", name)
}
