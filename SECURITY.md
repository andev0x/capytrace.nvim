# Security Policy

## Reporting Security Vulnerabilities

**Do not open a public issue for security vulnerabilities.**

If you discover a security vulnerability in capytrace.nvim, please report it responsibly by:

### 1. GitHub Security Advisory

Use [GitHub's private vulnerability reporting feature](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-about-security-advisories/privately-reporting-a-security-vulnerability):

1. Go to [Security tab](https://github.com/andev0x/capytrace.nvim/security/advisories)
2. Click "Report a vulnerability"
3. Fill out the form with:
   - Title and description
   - Vulnerability type
   - Affected versions
   - Proof of concept (if applicable)
   - Suggested fix (if you have one)

### 2. Email (Alternative)

Send an encrypted email to the maintainers with:

- Subject: `[SECURITY] Vulnerability Report`
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested remediation

**Note**: We're still setting up PGP keys for secure communication. For now, please use GitHub's private advisory feature.

## Response Timeline

We aim to:

- **Acknowledge receipt** within 48 hours
- **Provide initial assessment** within 5 business days
- **Release patch** within 2 weeks for critical vulnerabilities
- **Publicly disclose** once patch is released

## Vulnerability Disclosure Policy

### What We Consider a Security Vulnerability

- Code that allows unauthorized access
- Bypasses of security controls
- Leakage of sensitive information
- Denial of service attacks
- Privilege escalation
- Injection attacks
- Authentication/authorization bypass

### What We Don't Consider Critical

- Feature requests
- Documentation issues
- Configuration mistakes
- Expected behavior limitations
- Performance issues without security impact

## Security Best Practices

### For Users

1. **Keep capytrace updated**: Regularly update to the latest version
   ```bash
   # If installed via lazy.nvim, update regularly
   :Lazy update
   ```

2. **Review session data**: Check exported sessions for sensitive information
   - Remove credentials or secrets before sharing
   - Be cautious with terminal command logs

3. **Secure your data directory**:
   ```bash
   # Set proper permissions on capytrace data
   chmod 700 ~/.local/share/capytrace/
   ```

4. **Use strong machine authentication**: Ensure your workstation is secure

5. **Be cautious with shared sessions**: Don't share session files with untrusted parties

### For Developers

1. **Dependency scanning**: We use Go modules with security scanning
   ```bash
   go list -json -m all | nancy sleuth
   ```

2. **Code review**: All changes go through peer review

3. **Automated testing**: Security-relevant code has comprehensive tests

4. **Minimal dependencies**: We use only essential dependencies
   - Current: `modernc.org/sqlite` (pure Go, audited)

5. **Secure defaults**: All defaults prioritize security over features

## Known Issues

### Current Vulnerabilities

None known. Check [GitHub Security tab](https://github.com/andev0x/capytrace.nvim/security/advisories) for latest status.

### Security Limitations

1. **Local-Only Enforcement**: capytrace is designed for local use
   - No network communication (good for privacy)
   - Relies on OS file permissions
   - Multiple users on same machine can access each other's sessions

2. **No Encryption**: Session files are stored unencrypted
   - Recommendation: Store on encrypted disk
   - Use filesystem encryption (LUKS, FileVault, BitLocker)

3. **Terminal Commands**: Recorded as plaintext
   - Avoid recording sessions with sensitive commands
   - Remove credentials from session files before sharing

## Dependencies Security

### Current Dependencies

| Package | Version | License | Audited |
|---------|---------|---------|---------|
| modernc.org/sqlite | 1.44.0+ | Apache 2.0 | ✅ Yes |

### Dependency Updates

- Automatically checked via GitHub Dependabot
- Security patches applied promptly
- Monthly review of all dependencies

### Reporting Dependency Vulnerabilities

If you find a vulnerability in a dependency:

1. Report to the dependency maintainers first
2. Once patched, we'll update our version
3. We'll release a new version within 24 hours

## Security Testing

We perform:

- **Static Analysis**: `go vet`, `golangci-lint`
- **Race Detection**: `go test -race ./...`
- **Dependency Scanning**: Nancy, Dependabot
- **Code Review**: All changes reviewed by maintainers

## Version Support

### Supported Versions

| Version | Go Version | Neovim Version | Support |
|---------|-----------|-----------------|---------|
| 0.2.0+ | 1.18+ | 0.9.0+ | ✅ Current |
| 0.1.x | 1.18+ | 0.9.0+ | ⚠️ Limited |

### Support Timeline

- **Latest version**: Full support and security updates
- **Previous version**: Security updates for 6 months
- **Older versions**: No guaranteed support

## Security Roadmap

### Planned Improvements

- [ ] Optional session encryption
- [ ] File integrity checking
- [ ] Automatic credential redaction
- [ ] Audit logging for shared installations
- [ ] Security hardening guide

### Community Feedback

We welcome suggestions for security improvements:
- Open a discussion in [GitHub Discussions](https://github.com/andev0x/capytrace.nvim/discussions)
- Include threat model and use case
- Suggest mitigation strategies

## Privacy Policy

### What Data We Collect

**We collect NOTHING.**

- No telemetry
- No usage tracking
- No error reporting to external services
- All data stays on your machine

### Data Stored Locally

capytrace stores:
- Session files (JSON or SQLite)
- User configuration
- Event logs from your editor

All stored in: `~/.local/share/capytrace/`

### Data You Control

You have complete control:
- Delete sessions anytime
- Export data in standard formats
- Move data to different machines
- Archive or back up as needed

### No Third-Party Access

We never:
- Send data to external servers
- Share data with third parties
- Collect analytics or telemetry
- Use tracking services

## Compliance

### Standards We Follow

- **OWASP Top 10**: Secure coding practices
- **Go Security**: Best practices from Go team
- **MIT License**: Clear licensing and terms

### User Rights

- Right to privacy ✅
- Right to data ownership ✅
- Right to deletion ✅
- Right to transparency ✅

## Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Guidelines](https://golang.org/doc/security)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)

## Contact

### Security Team

- **Primary Contacts**: Project maintainers
- **Response Time**: 48 hours or less
- **Preferred Method**: GitHub Private Advisory

### Transparency

- Vulnerability disclosures made public once patched
- Security advisories published on GitHub
- Acknowledgment of researchers (if desired)

---

**Last Updated**: January 14, 2025

For more information, see [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
