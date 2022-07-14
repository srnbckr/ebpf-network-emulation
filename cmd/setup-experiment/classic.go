package main

import (
	"os/exec"

	"github.com/pkg/errors"
)

func removeRootQDisc(allowFail bool) error {
	// tc qdisc del dev [TAP_NAME] root
	cmd := exec.Command(TC, "qdisc", "del", "dev", "br0", "root")

	if out, err := cmd.CombinedOutput(); !allowFail && err != nil {
		return errors.Wrapf(err, "%#v: output: %s", cmd.Args, out)
	}

	return nil
}
