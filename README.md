#dnd

Is a simple CLI for restrictions access to websites and applications running on an OS.
Currently only darwin is supported.

# Example

Example blocking Spotify, intentionally misstyped.

```bash
dnd add application "spoyqwe.app"
found 16 matches based on provided pattern "spoyqwe.app"
which one do you want to proceed with:
[1]	Spotify.app | /Applications
[2]	zoom.us.app | /Applications
[3]	Discord.app | /Applications
[4]	Numbers.app | /Applications
[5]	swagger.zip | /Applications/VMware Fusion.app/Contents/Public
[6]	Blender.app | /Applications
[7]	swagger_to_py_client.py | /opt/homebrew/bin
[8]	spfquery5.34 | /usr/bin
[9]	swagger_to_go_server.py | /opt/homebrew/bin
[10]	softwareupdate | /usr/sbin
[11]	swagger_to_elm_client.py | /opt/homebrew/bin
[12]	swagger_style.py | /opt/homebrew/bin
[13]	syscallbyproc.d | /usr/bin
[14]	iMovie.app | /Applications
[15]	SvtAv1EncApp | /opt/homebrew/bin
[16]	spoyqwe.app | general purporse pattern match
choose number between [1, 16]: 1
option 1 will be used to kill any processes that contains the given pattern "/Applications/Spotify.app"
processed 1 items
```

After updating the settings, they need to be commited to take effect.

```bash
dnd commit
Password:
Domains:
~ matched [0]
+ add [0]
- delete [0]

Applications
~ matched [0]
+ add [1]
	Pattern:/Applications/Spotify.app
- delete [0]

commit ? (yes/no): yes
```
