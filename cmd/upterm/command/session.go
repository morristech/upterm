package command

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/jingweno/upterm/host"
	"github.com/jingweno/upterm/host/api"
	"github.com/jingweno/upterm/host/api/swagger/models"
	"github.com/jingweno/upterm/upterm"
	"github.com/jingweno/upterm/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	flagAdminSocket string
)

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session",
		Aliases: []string{"se"},
		Short:   "Display session",
	}
	cmd.AddCommand(current())
	cmd.AddCommand(list())
	cmd.AddCommand(show())

	return cmd
}

func list() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List shared sessions",
		Long:    `List shared sessions. Session admin sockets are located in ~/.upterm.`,
		Example: `  # List shared sessions
  upterm session list`,
		RunE: listRunE,
	}

	return cmd
}

func show() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"i"},
		Short:   "Display session by name",
		Long:    `Display session by name.`,
		Example: `  # Display session by name
  upterm session info NAME`,
		RunE: infoRunE,
	}

	return cmd
}

func current() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "current",
		Aliases: []string{"c"},
		Short:   "Display the current session",
		Long:    `Display the current session. By default, the command fetches the current session from the admin socket path defined in the UPTERM_ADMIN_SOCKET environment variable. The UPTERM_ADMIN_SOCKET environment variable is set after a session is shared with 'upterm host'.`,
		Example: `  # Display the current session defined in $UPTERM_ADMIN_SOCKET
  upterm session current
  # Display the current session with a custom path
  upterm session current --admin-socket ADMIN_SOCKET_PATH`,
		PreRunE: validateCurrentRequiredFlags,
		RunE:    currentRunE,
	}

	cmd.PersistentFlags().StringVarP(&flagAdminSocket, "admin-socket", "", currentAdminSocketFile(), "admin unix domain socket (required)")

	return cmd
}

func listRunE(c *cobra.Command, args []string) error {
	uptermDir, err := utils.UptermDir()
	if err != nil {
		return err
	}

	sessions, err := listSessions(uptermDir)
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No session is found. Create one with `upterm host`.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Current", "Session", "Command", "Force Command", "Host"})
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetCenterSeparator("|")
	table.AppendBulk(sessions)

	table.Render()
	return nil
}

func infoRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing session name")
	}

	uptermDir, err := utils.UptermDir()
	if err != nil {
		return err
	}

	return displaySession(host.AdminSocketFile(uptermDir, args[0]))
}

func currentRunE(c *cobra.Command, args []string) error {
	return displaySession(flagAdminSocket)
}

func listSessions(dir string) ([][]string, error) {
	result := make([][]string, 0)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	currentAdminSocket := currentAdminSocketFile()
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		adminSocket := host.AdminSocketFile(dir, file.Name())
		session, err := session(adminSocket)
		if err != nil {
			continue
		}

		var current string
		if adminSocket == currentAdminSocket {
			current = "*"
		}

		result = append(
			result,
			[]string{
				current,
				session.SessionID,
				strings.Join(session.Command, " "),
				naIfEmpty(strings.Join(session.ForceCommand, " ")),
				session.Host,
			})
	}

	return result, nil
}

func displaySession(path string) error {
	session, err := session(path)
	if err != nil {
		return err
	}

	user, err := api.EncodeIdentifierSession(session)
	if err != nil {
		return err
	}

	host, port, err := net.SplitHostPort(session.Host)
	if err != nil {
		return err
	}

	sshCmd := fmt.Sprintf("ssh %s@%s", user, host)
	if port != "22" {
		sshCmd = fmt.Sprintf("%s -p %s", sshCmd, port)
	}

	data := [][]string{
		[]string{"Command:", strings.Join(session.Command, " ")},
		[]string{"Force Command:", naIfEmpty(strings.Join(session.ForceCommand, " "))},
		[]string{"Host:", session.Host},
		[]string{"SSH Session:", sshCmd},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"=== " + session.SessionID})
	table.SetHeaderLine(false)
	table.SetColWidth(len(sshCmd)) // set to the longest text
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("")
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()

	return nil
}

func currentAdminSocketFile() string {
	return os.Getenv(upterm.HostAdminSocketEnvVar)
}

func session(adminSocket string) (*models.APIGetSessionResponse, error) {
	client := host.AdminClient(adminSocket)
	resp, err := client.GetSession(nil)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload(), nil
}

func validateCurrentRequiredFlags(c *cobra.Command, args []string) error {
	missingFlagNames := []string{}
	if flagAdminSocket == "" {
		missingFlagNames = append(missingFlagNames, "admin-socket")
	}

	if len(missingFlagNames) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missingFlagNames, ", "))
	}

	return nil
}

func naIfEmpty(s string) string {
	if s == "" {
		return "n/a"
	}

	return s
}