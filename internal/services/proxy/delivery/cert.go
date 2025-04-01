package httpdelivery

import (
	"bytes"
	"fmt"
	"math/big"
	"os/exec"
)

func genCert(host string, serial *big.Int) ([]byte, error) {
	cmd := exec.Command("scripts/gen_cert.sh", host, fmt.Sprintf("%d", serial))

	var certOut, errOut bytes.Buffer
	cmd.Stdout = &certOut
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("certificate generation failed: %w, stderr: %s", err, errOut.String())
	}

	return certOut.Bytes(), nil
}
