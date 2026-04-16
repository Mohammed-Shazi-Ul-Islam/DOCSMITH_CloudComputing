# 🚀 Quick Start Guide - Docksmith Demo

## For Faculty Presentation

---

## Step 1: Switch to WSL in VS Code

1. Press `F1` or `Ctrl+Shift+P`
2. Type: `WSL: Reopen Folder in WSL`
3. Wait for VS Code to reload
4. Verify bottom-left shows: `WSL: Ubuntu`

---

## Step 2: Open Terminal and Navigate

```bash
# Terminal should already be in project directory
# If not:
cd DOCSMITH_F

# Verify you're in the right place
ls
# Should see: docksmith/ sample_app/ setup_base_image.py etc.
```

---

## Step 3: Setup Base Image (One-time)

```bash
python3 setup_base_image.py --image alpine:latest --rootfs-tar alpine-minirootfs-3.19.1-x86_64.tar.gz
```

**Expected:** `Imported base image alpine:latest (sha256:...)`

---

## Step 4: Demo Sequence

### 4.1 Cold Build (Cache Misses)

```bash
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Point out:**
- COPY shows `[CACHE MISS] 0.12s`
- RUN shows `[CACHE MISS] 0.34s`
- Total time: ~0.5s

### 4.2 Warm Build (Cache Hits)

```bash
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Point out:**
- COPY shows `[CACHE HIT]`
- RUN shows `[CACHE HIT]`
- Total time: ~0.03s (near-instant!)
- Same digest (reproducible builds)

### 4.3 List Images

```bash
python3 -m docksmith.main images
```

**Point out:**
- Shows alpine (1 layer) and myapp (3 layers)
- Digest, created timestamp

### 4.4 Run Container

```bash
python3 -m docksmith.main run myapp:latest
```

**Point out:**
- Script executes
- ENV variable injected
- File from RUN step is accessible

### 4.5 ENV Override

```bash
python3 -m docksmith.main run -e APP_NAME=Demo myapp:latest
```

**Point out:**
- APP_NAME changed to "Demo"

### 4.6 Cache Invalidation

```bash
# Edit source file
echo "# comment" >> sample_app/run.sh

# Rebuild
python3 -m docksmith.main build -t myapp:latest sample_app/
```

**Point out:**
- COPY shows `[CACHE MISS]` (file changed)
- RUN also shows `[CACHE MISS]` (cascade!)
- Different digest

```bash
# Restore file
git checkout sample_app/run.sh
```

### 4.7 Remove Image

```bash
python3 -m docksmith.main rmi myapp:latest
```

**Point out:**
- Manifest deleted
- Exclusive layers removed

---

## Your Explanation Points (Shazi)

When explaining your module (builder.py):

### 1. Parse Phase
"First, we parse the Docksmithfile using `parse_docksmithfile()` which validates each instruction and returns structured Instruction objects."

### 2. Build State
"We initialize build state - layers list, env_dict, workdir, cmd - and create a CacheManager instance."

### 3. FROM Instruction
"FROM loads the base image manifest and inherits its layers, ENV, WORKDIR, and CMD. This sets up our starting point."

### 4. State Instructions
"WORKDIR, ENV, and CMD only update state - they don't produce layers. This keeps the image config lightweight."

### 5. COPY Instruction
"For COPY, we compute a cache key from the instruction, state, and source file hashes. On cache miss, `_execute_copy()` assembles the filesystem, copies files, and creates a delta tar of only the copied files."

### 6. RUN Instruction
"For RUN, we take a filesystem snapshot before execution, run the command using Pranav's isolation primitive, take another snapshot, and create a layer from only the changed files."

### 7. Manifest Creation
"Finally, we build an ImageConfig from accumulated state, create an ImageManifest with all layers, compute its SHA-256 digest, and save it using Piyush's state module."

---

## Key Talking Points

1. **"The builder orchestrates all modules"**
   - Parser → Cache → Layers → Runtime → State

2. **"COPY and RUN produce layers through delta computation"**
   - Only changed files go in the layer
   - Keeps layers small and efficient

3. **"Cache is wired through deterministic key computation"**
   - Same inputs = same cache key = same digest
   - Cache miss triggers cascade

4. **"Helper functions enable clean separation"**
   - `_assemble_rootfs()` - used by both COPY and RUN
   - `_execute_copy()` - COPY-specific logic
   - `_execute_run()` - RUN-specific logic
   - `_serialize_env()` - deterministic ENV serialization

---

## If Faculty Asks...

**Q: How do you ensure reproducible builds?**
A: "In `layers.py`, we sort all tar entries lexicographically and zero all timestamps. This ensures identical inputs always produce byte-identical digests."

**Q: How does cache invalidation work?**
A: "When any COPY or RUN step misses, we set `cache._force_miss = True`, which makes all subsequent lookups return None. This ensures consistency."

**Q: How does container isolation work?**
A: "Pranav's `runtime.py` uses Linux namespaces - `unshare()` with `CLONE_NEWPID`, `CLONE_NEWNS`, `CLONE_NEWUTS` - and `chroot()` to isolate the process. The same primitive is used for both RUN during build and `docksmith run` at runtime."

**Q: What happens if a RUN command fails?**
A: "We check the exit code. If non-zero, we raise a RuntimeError and stop the build immediately. The partial layer is not stored."

---

## Files to Reference

- `shazi_presentation_guide.md` - Your detailed explanation
- `TESTING_COMMANDS.md` - All test commands with expected outputs
- `WSL_SETUP.md` - WSL setup instructions

---

**You're ready! Good luck with your presentation! 🎓🚀**
