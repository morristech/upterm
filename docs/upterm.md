## upterm

Secure Terminal Sharing

### Synopsis

Upterm is an open-source solution for sharing terminal sessions instantly with the public internet over secure tunnels.

### Examples

```
  # Host a terminal session by running $SHELL
  # The client's input/output is attached to the host's
  $ upterm host

  # Display the ssh connection string
  $ upterm session current
  === BO6NOSSTP9LL08DOQ0RG
  Command:                /bin/bash
  Force Command:          n/a
  Host:                   uptermd.upterm.dev:22
  SSH Session:            ssh bo6nosstp9ll08doq0rg:MTAuMC4xNzAuMTY0OjIy@uptermd.upterm.dev

  # Open a new terminal and connect to the session
  $ ssh bo6nosstp9ll08doq0rg:MTAuMC4xNzAuMTY0OjIy@uptermd.upterm.dev
```

### Options

```
  -h, --help   help for upterm
```

### SEE ALSO

* [upterm host](upterm_host.md)	 - Host a terminal session
* [upterm session](upterm_session.md)	 - Display session
* [upterm version](upterm_version.md)	 - Show version

###### Auto generated by spf13/cobra on 5-Jan-2020
