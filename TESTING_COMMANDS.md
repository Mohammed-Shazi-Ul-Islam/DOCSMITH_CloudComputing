# 🧪 Docksmith Testing Commands Guide
## Complete Feature Testing for Faculty Demo

---

## 📋 Prerequisites

**IMPORTANT: Use WSL (Windows Subsystem for Linux) for full functionality!**

Before running any commands:
1. Open WSL terminal in VS Code (or separate WSL terminal)
2. Navigate to your project directory
3. Ensure Python 3 is installed in WSL

### Quick WSL Setup:

```bash
# In VS Code, open WSL terminal (Ctrl+Shift+` then select WSL)
# Or click the bottom-left corner and select "Reopen in WSL"

# Navigate to your project (WSL can access Windows files)
cd /mnt/c/Users/YourUsername/path/to/DOCSMITH_F

# Or if already in WSL workspace, just:
cd DOCSMITH_F

# Verify Python is available
python3 --version
```

---

## 🚀 Step-by-Step Testing Guide

### Step 0: Setup Base Image (One-time only)

```bash
# Import the alpine base image
python3 setup_base_image.py --image alpine:latest --rootfs-tar alpine-minirootfs-3.19.1-x86_64.tar.gz
```

**Expected Output:**
```
Imported base image alpine:latest (sha256:abc123...)
```

**What this does:** 
- Extracts the Alpine rootfs (including symlinks - works perfectly in WSL!)
- Creates tar layer with all files
- Saves manifest to `~/.docksmith/images/alpine_latest.json`

---

### Step 1: Check Current Images (Should be empty or show only alpine)

```bash
python3 -m docksmith.main images
```

**Expected Output:**
```
NAME                 TAG        DIGEST          LAYERS  CREATED
alpine               latest     sha256:a1b2c3d  1       2024-...
```

---

### Step 2: First Build (Cold Build - All Cache Misses)

```bash
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Expected Output:**
```
Step 1/6 : FROM alpine:latest
Step 2/6 : WORKDIR /app
Step 3/6 : ENV APP_NAME=Docksmith
Step 4/6 : COPY . /app [CACHE MISS] 0.12s
Step 5/6 : RUN echo "build-ok" > /app/build.txt [CACHE MISS] 0.34s
Step 6/6 : CMD ["/bin/sh", "/app/run.sh"]

Successfully built sha256:... myapp:latest (0.52s)
```

**What to observe:**
- COPY and RUN show `[CACHE MISS]` with timing
- Total build time shown at the end
- WORKDIR, ENV, CMD don't show cache status (no layers)

---

### Step 3: Second Build (Warm Build - All Cache Hits)

```bash
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Expected Output:**
```
Step 1/6 : FROM alpine:latest
Step 2/6 : WORKDIR /app
Step 3/6 : ENV APP_NAME=Docksmith
Step 4/6 : COPY . /app [CACHE HIT]
Step 5/6 : RUN echo "build-ok" > /app/build.txt [CACHE HIT]
Step 6/6 : CMD ["/bin/sh", "/app/run.sh"]

Successfully built sha256:... myapp:latest (0.03s)
```

**What to observe:**
- COPY and RUN show `[CACHE HIT]` (no timing)
- Build is MUCH faster (near-instant)
- Same digest as first build (reproducible builds!)

---

### Step 4: List Images

```bash
python3 -m docksmith.main images
```

**Expected Output:**
```
NAME                 TAG        DIGEST          LAYERS  CREATED
alpine               latest     sha256:a1b2c3d  1       2024-...
myapp                latest     sha256:f6e5d4c  3       2024-...
```

**What to observe:**
- myapp has 3 layers (base + COPY + RUN)
- Shows name, tag, short digest, layer count, created timestamp

---

### Step 5: Run Container (Default CMD)

```bash
python3 -m docksmith.main run myapp:latest
```

**Expected Output:**
```
Hello from Docksmith sample app
APP_NAME=Docksmith
BUILD_MARKER=build-ok

[docksmith] Container exited with code: 0
```

**What to observe:**
- Script executes inside container
- ENV variable is injected
- File created during RUN is accessible
- Clean exit with code 0

---

### Step 6: Run Container with ENV Override

```bash
python3 -m docksmith.main run -e APP_NAME=CustomName myapp:latest
```

**Expected Output:**
```
Hello from Docksmith sample app
APP_NAME=CustomName
BUILD_MARKER=build-ok

[docksmith] Container exited with code: 0
```

**What to observe:**
- APP_NAME is overridden to "CustomName"
- `-e` flag works correctly

---

### Step 7: Run Container with Custom Command

```bash
python3 -m docksmith.main run myapp:latest /bin/sh -c "echo 'Custom command executed'"
```

**Expected Output:**
```
Custom command executed

[docksmith] Container exited with code: 0
```

**What to observe:**
- Custom command overrides default CMD
- Container executes the provided command

---

### Step 8: Test Cache Invalidation (Edit Source File)

```bash
# Edit the run.sh file
echo "echo 'Modified script'" >> sample_app/run.sh

# Rebuild
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Expected Output:**
```
Step 1/6 : FROM alpine:latest
Step 2/6 : WORKDIR /app
Step 3/6 : ENV APP_NAME=Docksmith
Step 4/6 : COPY . /app [CACHE MISS] 0.11s
Step 5/6 : RUN echo "build-ok" > /app/build.txt [CACHE MISS] 0.33s
Step 6/6 : CMD ["/bin/sh", "/app/run.sh"]

Successfully built sha256:... myapp:latest (0.49s)
```

**What to observe:**
- COPY shows `[CACHE MISS]` (source file changed)
- RUN also shows `[CACHE MISS]` (cascade behavior)
- Different digest than before

```bash
# Restore the file
git checkout sample_app/run.sh
```

---

### Step 9: Test --no-cache Flag

```bash
python3 -m docksmith.main build -t myapp:latest sample_app/ --no-cache
```

**Expected Output:**
```
Step 1/6 : FROM alpine:latest
Step 2/6 : WORKDIR /app
Step 3/6 : ENV APP_NAME=Docksmith
Step 4/6 : COPY . /app [CACHE MISS] 0.12s
Step 5/6 : RUN echo "build-ok" > /app/build.txt [CACHE MISS] 0.34s
Step 6/6 : CMD ["/bin/sh", "/app/run.sh"]

Successfully built sha256:... myapp:latest (0.52s)
```

**What to observe:**
- All steps show `[CACHE MISS]` even though nothing changed
- `--no-cache` forces full rebuild

---

### Step 10: Build Another Image (Test Layer Reuse)

```bash
# Create a new Docksmithfile
cat > /tmp/test_docksmith/Docksmithfile << 'EOF'
FROM alpine:latest
WORKDIR /app
COPY . /app
RUN echo "different command" > /app/test.txt
CMD ["/bin/sh"]
EOF

# Create context directory
mkdir -p /tmp/test_docksmith
echo "test file" > /tmp/test_docksmith/test.txt

# Build
python3 -m docksmith.main build -t testapp:v1 /tmp/test_docksmith/
```

**Expected Output:**
```
Step 1/5 : FROM alpine:latest
Step 2/5 : WORKDIR /app
Step 3/5 : COPY . /app [CACHE MISS] 0.10s
Step 4/5 : RUN echo "different command" > /app/test.txt [CACHE MISS] 0.32s
Step 5/5 : CMD ["/bin/sh"]

Successfully built sha256:... testapp:v1 (0.48s)
```

---

### Step 11: Verify Layer Storage

```bash
# Check layers directory
ls -lh ~/.docksmith/layers/

# Check images directory
ls -lh ~/.docksmith/images/

# Check cache index
cat ~/.docksmith/cache/index.json | python3 -m json.tool | head -20
```

**What to observe:**
- Layers are stored as SHA-256 hex filenames
- Images are stored as `<name>_<tag>.json`
- Cache index maps cache keys to layer digests

---

### Step 12: Remove Image

```bash
python3 -m docksmith.main rmi myapp:latest
```

**Expected Output:**
```
Deleted: myapp:latest  (2 exclusive layers removed)
```

**What to observe:**
- Image manifest deleted
- Exclusive layers removed (not shared with other images)

---

### Step 13: Verify Removal

```bash
python3 -m docksmith.main images
```

**Expected Output:**
```
NAME                 TAG        DIGEST          LAYERS  CREATED
alpine               latest     sha256:a1b2c3d  1       2024-...
testapp              v1         sha256:...      3       2024-...
```

**What to observe:**
- myapp:latest is gone
- Other images remain

---

### Step 14: Test Error Handling (Invalid Instruction)

```bash
# Create invalid Docksmithfile
cat > /tmp/invalid_docksmith/Docksmithfile << 'EOF'
FROM alpine:latest
INVALID_INSTRUCTION test
EOF

mkdir -p /tmp/invalid_docksmith

# Try to build
python3 -m docksmith.main build -t invalid:test /tmp/invalid_docksmith/
```

**Expected Output:**
```
[PARSE ERROR] Line 2: Unknown instruction 'INVALID_INSTRUCTION'.
  Valid instructions: CMD, COPY, ENV, FROM, RUN, WORKDIR
  Got: 'INVALID_INSTRUCTION test'
```

**What to observe:**
- Clear error message with line number
- Lists valid instructions
- Build fails immediately

---

### Step 15: Test Missing Base Image

```bash
cat > /tmp/missing_base/Docksmithfile << 'EOF'
FROM nonexistent:latest
WORKDIR /app
EOF

mkdir -p /tmp/missing_base

python3 -m docksmith.main build -t test:missing /tmp/missing_base/
```

**Expected Output:**
```
[BUILD ERROR] FROM: Image 'nonexistent:latest' not found.
  Run: python3 setup_base_image.py
```

**What to observe:**
- Clear error about missing base image
- Helpful suggestion to run setup script

---

## 🎯 Quick Demo Sequence (For Faculty)

Run these commands in order for a complete demo:

```bash
# 1. Show empty state
python3 -m docksmith.main images

# 2. Cold build (show cache misses)
python3 -m docksmith.main build -t myapp:latest sample_app/

# 3. Warm build (show cache hits)
python3 -m docksmith.main build -t myapp:latest sample_app/

# 4. List images
python3 -m docksmith.main images

# 5. Run container
python3 -m docksmith.main run myapp:latest

# 6. Run with ENV override
python3 -m docksmith.main run -e APP_NAME=Demo myapp:latest

# 7. Edit file and rebuild (show cascade)
echo "# comment" >> sample_app/run.sh
python3 -m docksmith.main build -t myapp:latest sample_app/
git checkout sample_app/run.sh

# 8. Remove image
python3 -m docksmith.main rmi myapp:latest
```

---

## 📊 What Each Command Tests

| Command | Tests Feature |
|---------|---------------|
| `build` (first time) | Parser, builder, layer creation, cache miss |
| `build` (second time) | Cache hit, reproducible builds |
| `build` (after edit) | Cache invalidation, cascade behavior |
| `build --no-cache` | Cache bypass |
| `images` | Manifest listing, state management |
| `run` | Container runtime, isolation, ENV injection |
| `run -e` | ENV override |
| `run <cmd>` | CMD override |
| `rmi` | Image deletion, layer cleanup |

---

## 🐧 WSL-Specific Notes

### Why WSL is Better for This Project:

1. **Real Container Isolation**: The `runtime.py` uses Linux namespaces (`unshare()`, `chroot()`) which only work on Linux
2. **Symlinks Work**: Alpine Linux uses symlinks extensively (e.g., `/bin/sh` → `/bin/busybox`)
3. **Authentic Behavior**: This is how Docker actually works under the hood

### Testing Container Isolation (WSL Only):

```bash
# Run container and try to write a file
python3 -m docksmith.main run myapp:latest /bin/sh -c "echo 'test' > /tmp/isolated.txt"

# Check if file exists on host (it shouldn't!)
ls /tmp/isolated.txt
# Output: ls: cannot access '/tmp/isolated.txt': No such file or directory

# This proves isolation is working! ✅
```

### Checking Where Files Are Stored:

```bash
# In WSL, docksmith files are stored in your WSL home directory
ls -la ~/.docksmith/

# Output:
# drwxr-xr-x  images/
# drwxr-xr-x  layers/
# drwxr-xr-x  cache/
```

### If You Need to Access from Windows:

```bash
# WSL home directory is accessible from Windows at:
# \\wsl$\Ubuntu\home\yourusername\.docksmith\

# Or open in Windows Explorer:
explorer.exe ~/.docksmith/
```

---

## 🔍 Debugging Commands

If something goes wrong:

```bash
# Check if base image exists
ls -la ~/.docksmith/images/

# Check layer files
ls -la ~/.docksmith/layers/

# Check cache index
cat ~/.docksmith/cache/index.json

# View a specific manifest
cat ~/.docksmith/images/alpine_latest.json | python3 -m json.tool

# Clear cache (if needed)
rm -f ~/.docksmith/cache/index.json

# Full reset (nuclear option)
rm -rf ~/.docksmith/
```

---

## ✅ Expected Results Summary

After running all tests, you should have verified:

- [ ] Parser validates instructions correctly
- [ ] Builder orchestrates all modules
- [ ] Cache hit/miss logic works
- [ ] Cascade behavior on cache miss
- [ ] Reproducible builds (same digest on cache hit)
- [ ] Layer creation and storage
- [ ] Container runtime and isolation (WSL only!)
- [ ] ENV injection and override
- [ ] CMD execution
- [ ] Image listing
- [ ] Image removal
- [ ] Error handling

---

**All commands ready for your faculty demo! 🚀**
