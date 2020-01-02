## upterm session current

Display the current session

### Synopsis

Display the current session. By default, the command fetches the current session from the admin socket path defined in the UPTERM_ADMIN_SOCKET environment variable. The UPTERM_ADMIN_SOCKET environment variable is set after a session is shared with 'upterm host'.

```
upterm session current [flags]
```

### Examples

```
  # Display the current session defined in $UPTERM_ADMIN_SOCKET
  upterm session current
  # Display the current session with a custom path
  upterm session current --admin-socket ADMIN_SOCKET_PATH
```

### Options

```
      --admin-socket string   admin unix domain socket (required)
  -h, --help                  help for current
```

### SEE ALSO

* [upterm session](upterm_session.md)	 - Display session

###### Auto generated by spf13/cobra on 1-Jan-2020