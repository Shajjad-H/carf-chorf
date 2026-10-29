package report_service

import (
	"os"
	"testing"

	"carf-chorf.com/app/app/kernel"
	"github.com/google/uuid"
)

func TestSaveReadHas(t *testing.T) {
	previous := kernel.ReportStorage.Path
	kernel.ReportStorage.Path = t.TempDir()
	t.Cleanup(func() {
		kernel.ReportStorage.Path = previous
	})

	token := uuid.NewString()
	if Has(token) {
		t.Fatal("report unexpectedly exists")
	}
	if err := Save(token, []byte("<html>report</html>")); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if !Has(token) {
		t.Fatal("report was not saved")
	}
	data, err := Read(token)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(data) != "<html>report</html>" {
		t.Fatalf("report = %q", data)
	}

	path := Dir() + "/" + token + ".html"
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report file missing: %v", err)
	}

	if err := Delete(token); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if Has(token) {
		t.Fatal("report still exists after Delete()")
	}
	if err := Delete(token); err != nil {
		t.Fatalf("Delete() missing file error = %v", err)
	}
}

func TestInvalidToken(t *testing.T) {
	if err := Save("bad", []byte("data")); err == nil {
		t.Fatal("Save() accepted an invalid token")
	}
	if _, err := Read("bad"); err == nil {
		t.Fatal("Read() accepted an invalid token")
	}
	if Has("bad") {
		t.Fatal("Has() accepted an invalid token")
	}
}
