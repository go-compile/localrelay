# Security Policy

## **IMPORTANT**
- DO NOT OPEN A GITHUB ISSUE OR PULL REQUEST *(unless following reporting method B)*.
- REPORT IT PRIVATELY.
- PROVIDE REQUIRED DETAILS *(listed bellow)*.

## Supported Versions

Only the latest version is supported for security updates.

## Reporting a Vulnerability
When reporting a vulnerability (vuln) you **must** provide at-lest the following:
- Vulnerability type/classification
- Is it a dependency? (true/false)
- Affected component (source file, function or route)
- Impact (what happens when exploited)
- Justify importance (**only** if the vuln is obscure and has no [CWE](https://cwe.mitre.org/data/published/cwe_latest.pdf) identifier)
- Impacted version (version number or git commit hash)
- Impacted platforms (windows, mac, linux, openbsd etc)

**Recommended** including:
- Poof of Concept
- Explanation of attack and how to reproduce
- Patch

### CVEs
If you obtain a CVE for a vulnerability found in this repository please contact me with the CVE identifier.

### Reporting Method A
To report a vulnerability, navigate the Localrelay's Github repository. Here you will find a tab called "security", next privately submit a vulnerability via Github's built in system.

### Reporting Method B
Alternatively, find my contact details are (signed and) provided at <https://github.com/go-compile/public-key>. However, there is guarantee no I will see your message. Resulting, you are recommended to make a public issue which only asks for my contact information. **Do not:** (1) make a public issue disclosing the vuln, (2) make a public issue stating there is a vuln.

## Full-Disclosure
Full disclosure, opposed to responsible disclosure (privately reporting and awaiting a patch), is ill-advised. Upon notification of a vulnerability (in-private), a patch will be issued with upmost urgency in a timely manor, thus no need for full-disclosure. However, if you do opt for full-disclosure, please still contact me and provide access to the publication material.