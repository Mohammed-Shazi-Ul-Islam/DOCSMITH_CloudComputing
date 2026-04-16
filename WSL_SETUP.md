# 🐧 WSL Setup Guide for Docksmith

## Why Use WSL?

Docksmith uses Linux-specific features that only work properly in a Linux environment:
- **Namespaces** (`unshare()`, `CLONE_NEWPID`, `CLONE_NEWNS`, `CLONE_NEWUTS`)
- **chroot** for filesystem isolation
- **Symlinks** (Alpine Linux uses them extensively)

---

## Quick Start: Using WSL in VS Code

### Option 1: Reopen Folder in WSL (Recommended)

1. Open your project in VS Code
2. Press `F1` or `Ctrl+Shift+P`
3. Type: `WSL: Reopen Folder in WSL`
4. Select your WSL distribution (usually Ubuntu)
5. VS Code will reload with WSL backend

**You'll know you're in WSL when:**
- Bottom-left corner shows: `WSL: Ubuntu` (or your distro name)
- Terminal shows Linux prompt: `username@hostname:~$`

### Option 2: Open WSL Terminal in VS Code

1. Open terminal: `Ctrl+Shift+` ` (backtick)
2. Click the dropdown arrow next to `+`
3. Select your WSL distribution
4. Navigate to project: `cd /mnt/c/path/to/DOCSMITH_F`

---

## Running Commands in WSL

Once you're in WSL, all commands work exactly as documented:

```bash
# 1. Navigate to project (if not already there)
cd DOCSMITH_F

# 2. Setup base image
python3 setup_base_image.py --image alpine:latest --rootfs-tar alpine-minirootfs-3.19.1-x86_64.tar.gz

# 3. Build image
python3 -m docksmith.main build -t myapp:latest sample_app/

# 4. Run container
python3 -m docksmith.main run myapp:latest
```

---

## File Locations in WSL

### Where Docksmith Stores Files:

```bash
# In WSL, files are stored in your WSL home directory
~/.docksmith/
├── images/      # Image manifests
├── layers/      # Layer tar files
└── cache/       # Cache index
```

### Accessing WSL Files from Windows:

**Method 1: Windows Explorer**
```
\\wsl$\Ubuntu\home\yourusername\.docksmith\
```

**Method 2: From WSL Terminal**
```bash
# Open current directory in Windows Explorer
explorer.exe .

# Open docksmith directory
explorer.exe ~/.docksmith/
```

---

## Verifying WSL is Working

### Test 1: Check Linux Kernel

```bash
uname -a
# Should show: Linux ... Microsoft ...
```

### Test 2: Check Python

```bash
python3 --version
# Should show: Python 3.x.x
```

### Test 3: Test Namespace Support

```bash
# This should work in WSL (may need sudo)
python3 -c "import os; print(hasattr(os, 'unshare'))"
# Should print: True
```

---

## Common Issues & Solutions

### Issue 1: "python3: command not found"

```bash
# Install Python in WSL
sudo apt update
sudo apt install python3 python3-pip
```

### Issue 2: "Permission denied" when running setup

```bash
# Make script executable
chmod +x setup_base_image.py

# Or just use python3 directly
python3 setup_base_image.py --image alpine:latest --rootfs-tar alpine-minirootfs-3.19.1-x86_64.tar.gz
```

### Issue 3: Can't find project files

```bash
# Windows C: drive is mounted at /mnt/c in WSL
cd /mnt/c/Users/YourUsername/path/to/DOCSMITH_F

# Or if you reopened in WSL, you're already in the right place
pwd  # Check current directory
```

### Issue 4: "Operation not permitted" during container run

```bash
# Some namespace operations need elevated privileges
# Try with sudo (for testing only)
sudo python3 -m docksmith.main run myapp:latest

# Or run without full isolation (fallback mode)
# The code should handle this gracefully
```

---

## Best Practices for Faculty Demo

1. **Open VS Code in WSL mode** before the demo
2. **Test all commands beforehand** to ensure everything works
3. **Keep a WSL terminal open** during presentation
4. **Show the isolation test** to prove containers are actually isolated:
   ```bash
   # Write file in container
   python3 -m docksmith.main run myapp:latest /bin/sh -c "echo test > /tmp/demo.txt"
   
   # Verify it doesn't exist on host
   ls /tmp/demo.txt  # Should fail - file is isolated!
   ```

---

## Quick Reference: WSL vs Windows

| Feature | Windows (Native) | WSL |
|---------|-----------------|-----|
| Parse Docksmithfile | ✅ Works | ✅ Works |
| Build images | ✅ Works | ✅ Works |
| Cache system | ✅ Works | ✅ Works |
| Layer storage | ✅ Works | ✅ Works |
| Symlinks | ❌ Issues | ✅ Works |
| Container isolation | ❌ No namespaces | ✅ Full isolation |
| `chroot()` | ❌ Not available | ✅ Works |
| Runtime execution | ⚠️ Limited | ✅ Full support |

---

## Summary

**For your faculty presentation, use WSL!** It provides:
- ✅ Full Linux environment
- ✅ Real container isolation
- ✅ Authentic Docker-like behavior
- ✅ All features working as designed

**Commands to remember:**
```bash
# Switch to WSL in VS Code
F1 → "WSL: Reopen Folder in WSL"

# Navigate to project
cd DOCSMITH_F

# Run any command
python3 -m docksmith.main <command>
```

---

**You're all set for an impressive demo! 🚀**
