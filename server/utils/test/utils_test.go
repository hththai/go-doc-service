package test

import (
	"2_Go/utils"
	"testing"
)

func TestSanitizeFileName(t *testing.T) {
	pathName := `$%$Hayden Thai_Resume-2.pdf`

	checkName, err := utils.SanitizeFileName(pathName)

	if err != nil {
		t.Fatalf("Cannot Sanity file name %s", pathName)
	}

	t.Logf("GOOD sanitized path is : %s", checkName)
}
