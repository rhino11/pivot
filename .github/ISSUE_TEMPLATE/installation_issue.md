---
name: Installation Issue
about: Problems installing or updating Pivot CLI
title: '[INSTALL] '
labels: 'installation, help-wanted'
assignees: ''

---

## Installation Method
<!-- Please select the installation method that's having issues -->
- [ ] Homebrew (macOS): `brew install rhino11/tap/pivot`
- [ ] APT (Ubuntu/Debian): Download and install .deb package
- [ ] YUM/DNF (RHEL/Fedora): Download and install .rpm package
- [ ] Direct download: Manual binary installation
- [ ] Build from source: `go build` or `make build`

## Operating System
<!-- Please provide your OS details -->
- **OS**: <!-- e.g., macOS 14.1, Ubuntu 22.04, Windows 11 -->
- **Architecture**: <!-- e.g., AMD64, ARM64 -->

## Issue Description
<!-- A clear description of what went wrong -->

## Steps to Reproduce
<!-- Please provide the exact commands you ran -->
1. 
2. 
3. 

## Error Output
<!-- Please paste the complete error message -->
```
[Paste error output here]
```

## Expected Behavior
<!-- What did you expect to happen? -->

## Additional Context
<!-- For Homebrew issues -->
- [ ] I've run `brew update` before trying to install
- [ ] I've checked if the tap exists: `brew search rhino11/tap/pivot`
- [ ] I've tried the manual update: `brew untap rhino11/tap && brew tap rhino11/tap`

<!-- For manual installation -->
- [ ] I've verified the download integrity using the checksums
- [ ] I've made the binary executable: `chmod +x pivot`
- [ ] I've placed the binary in my PATH

## System Information (if relevant)
<!-- Please run and paste the output if Pivot is partially working -->
```bash
# If pivot is partially installed, run:
pivot version

# For Homebrew debugging:
brew doctor
brew config
```

## Workaround Found
<!-- If you found a workaround, please share it to help others -->
