package ftests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jingweno/upterm/host"
)

func testClientNonExistingSession(t *testing.T, testServer TestServer) {
	adminSockDir, err := newAdminSocketDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(adminSockDir)

	adminSocketFile := filepath.Join(adminSockDir, "upterm.sock")

	h := &Host{
		Command:                  []string{"bash", "-c", "PS1='' bash --norc"},
		PrivateKeys:              []string{HostPrivateKey},
		AdminSocketFile:          adminSocketFile,
		PermittedClientPublicKey: ClientPublicKeyContent,
	}
	if err := h.Share(testServer.Addr()); err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	// verify admin server
	adminClient := host.AdminClient(adminSocketFile)
	resp, err := adminClient.GetSession(nil)
	if err != nil {
		t.Fatal(err)
	}
	session := resp.GetPayload()
	if want, got := h.SessionID, session.SessionID; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.Addr(), session.Host; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.NodeAddr(), session.NodeAddr; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	// verify input/output
	hostInputCh, hostOutputCh := h.InputOutput()
	hostScanner := scanner(hostOutputCh)

	hostInputCh <- "echo hello"
	if want, got := "echo hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	c := &Client{
		PrivateKeys: []string{ClientPrivateKey},
	}
	session.SessionID = "not-existance" // set session ID to non-existance
	err = c.Join(session, testServer.Addr())

	// Unfortunately there is no explicit error to the client.
	// But ssh handshake fails with the connection closed
	if want, got := "ssh: handshake failed: EOF", err.Error(); want != got {
		t.Fatalf("Unexpected error, want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
}

func testClientAttachHostWithSameCommand(t *testing.T, testServer TestServer) {
	adminSockDir, err := newAdminSocketDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(adminSockDir)

	adminSocketFile := filepath.Join(adminSockDir, "upterm.sock")

	h := &Host{
		Command:                  []string{"bash", "-c", "PS1='' bash --norc"},
		PrivateKeys:              []string{HostPrivateKey},
		AdminSocketFile:          adminSocketFile,
		PermittedClientPublicKey: ClientPublicKeyContent,
	}
	if err := h.Share(testServer.Addr()); err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	// verify admin server
	adminClient := host.AdminClient(adminSocketFile)
	resp, err := adminClient.GetSession(nil)
	if err != nil {
		t.Fatal(err)
	}
	session := resp.GetPayload()
	if want, got := h.SessionID, session.SessionID; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.Addr(), session.Host; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.NodeAddr(), session.NodeAddr; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	// verify input/output
	hostInputCh, hostOutputCh := h.InputOutput()
	hostScanner := scanner(hostOutputCh)

	hostInputCh <- "echo hello"
	if want, got := "echo hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	c := &Client{
		PrivateKeys: []string{ClientPrivateKey},
	}
	if err := c.Join(session, testServer.Addr()); err != nil {
		t.Fatal(err)
	}

	remoteInputCh, remoteOutputCh := c.InputOutput()
	remoteScanner := scanner(remoteOutputCh)

	// remote stdout should receive the last output of host when joining
	if want, got := "hello", scan(remoteScanner); want != got {
		// HACK: sometimes the last output include more than the previous line.
		// Try the next line if that's the case
		if want, got := "hello", scan(remoteScanner); want != got {
			t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
		}
	}

	remoteInputCh <- "echo hello again"
	if want, got := "echo hello again", scan(remoteScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello again", scan(remoteScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}

	// host should link to remote with the same input/output
	if want, got := "echo hello again", scan(hostScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello again", scan(hostScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}
}

func testClientAttachHostWithDifferentCommand(t *testing.T, testServer TestServer) {
	adminSockDir, err := newAdminSocketDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(adminSockDir)

	adminSocketFile := filepath.Join(adminSockDir, "upterm.sock")

	h := &Host{
		Command:                  []string{"bash", "-c", "PS1='' bash --norc"},
		ForceCommand:             []string{"bash", "-c", "PS1='' bash --norc"},
		PrivateKeys:              []string{HostPrivateKey},
		AdminSocketFile:          adminSocketFile,
		PermittedClientPublicKey: ClientPublicKeyContent,
	}
	if err := h.Share(testServer.Addr()); err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	// verify admin server
	adminClient := host.AdminClient(adminSocketFile)
	resp, err := adminClient.GetSession(nil)
	if err != nil {
		t.Fatal(err)
	}
	session := resp.GetPayload()
	if want, got := h.SessionID, session.SessionID; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.Addr(), session.Host; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := testServer.NodeAddr(), session.NodeAddr; want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	// verify input/output
	hostInputCh, hostOutputCh := h.InputOutput()
	hostScanner := scanner(hostOutputCh)

	hostInputCh <- "echo hello"
	if want, got := "echo hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}

	c := &Client{
		PrivateKeys: []string{ClientPrivateKey},
	}
	if err := c.Join(session, testServer.Addr()); err != nil {
		t.Fatal(err)
	}

	remoteInputCh, remoteOutputCh := c.InputOutput()
	remoteScanner := scanner(remoteOutputCh)
	time.Sleep(1 * time.Second) // HACK: wait for ssh stdin/stdout to fully attach

	remoteInputCh <- "echo hello again"
	if want, got := "echo hello again", scan(remoteScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello again", scan(remoteScanner); want != got {
		t.Fatalf("want=%q got=%q:\n%s", want, got, cmp.Diff(want, got))
	}

	// host shouldn't be linked to remote
	hostInputCh <- "echo hello"
	if want, got := "echo hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
	if want, got := "hello", scan(hostScanner); want != got {
		t.Fatalf("want=%s got=%s:\n%s", want, got, cmp.Diff(want, got))
	}
}

func newAdminSocketDir() (string, error) {
	return ioutil.TempDir("", "upterm")
}
