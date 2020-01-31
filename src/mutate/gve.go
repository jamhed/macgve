package mutate

import "strings"

type GoVaultEnv struct {
	image      string
	vaultaddr  string
	authpath   string
	containers []string
}

func (gve *GoVaultEnv) IsIn(container string) bool {
	if len(gve.containers) == 0 {
		return true
	}
	for _, c := range gve.containers {
		if c == container {
			return true
		}
	}
	return false
}

func (gve *GoVaultEnv) SetContainers(containers string) {
	gve.containers = strings.Split(containers, ",")
}
